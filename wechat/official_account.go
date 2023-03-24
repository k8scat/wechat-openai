package wechat

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	oaConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"go.uber.org/zap"

	"github.com/k8scat/wechat-openai/config"
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

	act, _ := server.GetAccessToken()
	http.Get(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/menu/delete?access_token=%s", act))

	// 设置接收消息的处理方法
	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		openaiClient := openai.GetClient()
		resp, err := openai.Chat(openaiClient, msg.Content)
		if err != nil {
			log.Error("openai chat failed", zap.Error(err), zap.Stack("stack"), zap.Any("wechat_msg", msg), zap.Any("openai_response", resp))
			return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText("openai chat failed")}
		}

		log.Info("openai chat success", zap.Any("wechat_msg", msg), zap.Any("openai_response", resp))
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: message.NewText(resp.Choices[0].Message.Content)}
	})

	// 处理消息接收以及回复
	err := server.Serve()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// 发送回复的消息
	return server.Send()

}
