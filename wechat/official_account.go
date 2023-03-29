package wechat

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	oaConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"go.uber.org/zap"

	"github.com/k8scat/wechat-openai/config"
	"github.com/k8scat/wechat-openai/db"
	"github.com/k8scat/wechat-openai/log"
	"github.com/k8scat/wechat-openai/openai"
)

var (
	initOfficialAccount sync.Once
	officialAccount     *officialaccount.OfficialAccount
)

func GetOfficialAccount() *officialaccount.OfficialAccount {
	initOfficialAccount.Do(func() {
		// 使用memcache保存access_token，也可选择redis或自定义cache
		wc := wechat.NewWechat()
		memory := cache.NewMemory()
		c := config.GetConfig()
		cfg := &oaConfig.Config{
			AppID:          c.Wechat.AppID,
			AppSecret:      c.Wechat.AppSecret,
			Token:          c.Wechat.Token,
			EncodingAESKey: c.Wechat.EncodingAESKey,
			Cache:          memory,
		}
		officialAccount = wc.GetOfficialAccount(cfg)
	})
	return officialAccount
}

func HandleMessage(oa *officialaccount.OfficialAccount, req *http.Request, w http.ResponseWriter) error {
	server := oa.GetServer(req, w)

	// 设置接收消息的处理方法
	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		msgID := fmt.Sprintf("%d", msg.MsgID)

		ctx := context.Background()
		conn := db.GetRedisClient().Conn()
		defer conn.Close()

		if conn.Exists(ctx, msgID).Val() > 0 {
			log.Warn("duplicated msg", zap.Any("wechat_msg", msg))
			return nil
		}

		if err := conn.SetEx(ctx, msgID, "", time.Minute).Err(); err != nil {
			log.Error("cache msg id failed", zap.Error(err), zap.Stack("stack"), zap.Any("wechat_msg", msg))
			return nil
		}

		if msg.Content == "" {
			log.Info("empty content", zap.Any("wechat_msg", msg))
			return nil
		}

		ch := make(chan interface{}, 1)
		go func() {
			openaiClient := openai.GetClient()
			resp, err := openai.Chat(openaiClient, msg.Content)
			if err != nil {
				ch <- errors.Trace(err)
				return
			}
			ch <- resp
		}()

		f := func(v interface{}) *openai.ChatCompletionResponse {
			switch v.(type) {
			case error:
				log.Error("openai chat failed", zap.Error(v.(error)), zap.Stack("stack"), zap.Any("wechat_msg", msg))
			case *openai.ChatCompletionResponse:
				return v.(*openai.ChatCompletionResponse)
			}
			return nil
		}

		var resp *openai.ChatCompletionResponse
		select {
		case v := <-ch:
			resp = f(v)
		case <-time.After(4 * time.Second):
			log.Warn("openai chat timeout", zap.Any("wechat_msg", msg))

			go func() {
				var resp *openai.ChatCompletionResponse
				select {
				case v := <-ch:
					resp = f(v)
					if resp == nil {
						return
					}
				}

				conn := db.GetRedisClient().Conn()
				defer conn.Close()

				err := conn.SetEx(ctx, msgID, resp.Choices[0].Message.Content, time.Minute*5).Err()
				if err != nil {
					log.Error("cache message response failed", zap.Error(err), zap.Stack("stack"), zap.Any("wechat_msg", msg), zap.Any("openai_response", resp))
				}
			}()

			url := fmt.Sprintf("%s/message/%d", config.GetConfig().App.BaseURL, msg.MsgID)
			return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(url)}
		}

		if resp == nil {
			return nil
		}

		log.Info("openai chat success", zap.Any("wechat_msg", msg), zap.Any("openai_response", resp))
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(resp.Choices[0].Message.Content)}
	})

	// 处理消息接收以及回复
	err := server.Serve()
	if err != nil {
		log.Error("serve failed", zap.Error(err), zap.Stack("stack"))
		return nil
	}
	// 发送回复的消息
	return server.Send()

}
