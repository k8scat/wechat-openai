package openai

import (
	"context"
	"sync"

	"github.com/juju/errors"
	"github.com/sashabaranov/go-openai"

	"github.com/k8scat/wechat-openai/config"
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

func Chat(client *openai.Client, content string) (*openai.ChatCompletionResponse, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &resp, nil
}
