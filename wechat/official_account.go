package wechat

import (
	"context"
	"encoding/json"
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
		c := config.GetConfig()
		cfg := &oaConfig.Config{
			AppID:          c.Wechat.AppID,
			AppSecret:      c.Wechat.AppSecret,
			Token:          c.Wechat.Token,
			EncodingAESKey: c.Wechat.EncodingAESKey,
			Cache:          cache.NewMemory(),
		}
		officialAccount = wc.GetOfficialAccount(cfg)
	})
	return officialAccount
}

func HandleMessage(oa *officialaccount.OfficialAccount, req *http.Request, w http.ResponseWriter) error {
	server := oa.GetServer(req, w)

	// 设置接收消息的处理方法
	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		if msg.Content == "" {
			log.Info("empty content", zap.Any("wechat_msg", msg))
			return nil
		}

		userID := string(msg.FromUserName)
		msgKey := fmt.Sprintf("message:%d", msg.MsgID)

		conn := db.GetRedisClient().Conn()
		defer conn.Close()
		ctx := context.Background()

		if msg.Content == "/help" {
			return &message.Reply{
				MsgType: message.MsgTypeText,
				MsgData: message.NewText("chat with me, send /clear to clear chat history"),
			}
		} else if msg.Content == "/clear" {
			if err := conn.Del(ctx, userID).Err(); err != nil {
				log.Error("clear chat history failed", zap.Error(err), zap.Stack("stack"), zap.Any("wechat_msg", msg))
			}
			return &message.Reply{
				MsgType: message.MsgTypeText,
				MsgData: message.NewText("chat history cleared"),
			}
		}

		if conn.HExists(ctx, userID, msgKey).Val() {
			log.Warn("duplicated msg", zap.Any("wechat_msg", msg))
			return nil
		}

		chat := &openai.Chat{
			Question: msg.Content,
		}
		b, err := json.Marshal(chat)
		if err != nil {
			log.Error("marshal failed", zap.Error(err), zap.Stack("stack"), zap.Any("wechat_msg", msg))
			return nil
		}
		if err := conn.HSet(ctx, userID, msgKey, string(b)).Err(); err != nil {
			log.Error("store chat failed", zap.Error(err), zap.Stack("stack"), zap.Any("wechat_msg", msg))
			return nil
		}

		ch := make(chan interface{}, 1)
		go func() {
			openaiClient := openai.GetClient()
			resp, err := openai.CreateChat(openaiClient, userID, msg.Content)
			if err != nil {
				ch <- errors.Trace(err)
				return
			}
			ch <- resp
		}()

		f := func(v interface{}) *openai.ChatCompletionResponse {
			switch v.(type) {
			case error:
				log.Error("chat failed", zap.Error(v.(error)), zap.Stack("stack"), zap.Any("wechat_msg", msg))
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
			log.Warn("chat timeout", zap.Any("wechat_msg", msg))

			go func(userID, msgKey string, t *openai.Chat) {
				var resp *openai.ChatCompletionResponse
				select {
				case v := <-ch:
					resp = f(v)
					if resp == nil {
						return
					}
				}

				t.Answer = resp.Choices[0].Message.Content
				b, err := json.Marshal(t)
				if err != nil {
					log.Error("marshal failed", zap.Error(err), zap.Stack("stack"), zap.Any("wechat_msg", msg), zap.Any("openai_response", resp))
					return
				}

				conn := db.GetRedisClient().Conn()
				defer conn.Close()
				ctx := context.Background()
				if err := conn.HSet(ctx, userID, msgKey, string(b)).Err(); err != nil {
					log.Error("store chat failed", zap.Error(err), zap.Stack("stack"), zap.Any("wechat_msg", msg), zap.Any("openai_response", resp))
				}
			}(userID, msgKey, chat)

			url := fmt.Sprintf("%s/user/%s/message/%d", config.GetConfig().App.BaseURL, userID, msg.MsgID)
			return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(url)}
		}

		if resp == nil {
			return nil
		}

		log.Info("chat success", zap.Any("wechat_msg", msg), zap.Any("openai_response", resp))
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
