package openai

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/juju/errors"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"

	"github.com/k8scat/wechat-openai/config"
	"github.com/k8scat/wechat-openai/db"
	"github.com/k8scat/wechat-openai/log"
)

type ChatCompletionResponse = openai.ChatCompletionResponse

var (
	initClient sync.Once
	client     *openai.Client
)

func GetClient() *openai.Client {
	initClient.Do(func() {
		cfg := config.GetConfig()
		clientCfg := openai.DefaultConfig(cfg.OpenAI.Key)
		if cfg.OpenAI.BaseURL != "" {
			clientCfg.BaseURL = cfg.OpenAI.BaseURL
		}
		client = openai.NewClientWithConfig(clientCfg)
		client = openai.NewClient(cfg.OpenAI.Key)
	})
	return client
}

type Chat struct {
	Question string `json:"question,omitempty"`
	Answer   string `json:"answer,omitempty"`
}

func CreateChat(client *openai.Client, userID, content string) (*openai.ChatCompletionResponse, error) {
	msgs, err := loadChatContext(userID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: msgs,
		},
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &resp, nil
}

func loadChatContext(userID string) ([]openai.ChatCompletionMessage, error) {
	conn := db.GetRedisClient().Conn()
	defer conn.Close()
	cmd := conn.HGetAll(context.Background(), userID)
	if cmd.Err() != nil {
		return nil, errors.Trace(cmd.Err())
	}
	talks, err := cmd.Result()
	if err != nil {
		return nil, errors.Trace(err)
	}
	msgs := make([]openai.ChatCompletionMessage, 0, len(talks)*2+1)
	for _, v := range talks {
		var chat Chat
		if err := json.Unmarshal([]byte(v), &chat); err != nil {
			log.Error("unmarshal failed", zap.Error(err), zap.Stack("stack"), zap.String("chat", v))
			continue
		}

		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: chat.Question,
		}, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: chat.Answer,
		})
	}
	return msgs, nil
}
