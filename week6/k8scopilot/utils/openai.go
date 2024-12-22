package utils

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	Client *openai.Client
	ctx    context.Context
}

func NewOpenAIClient() (*OpenAI, error) {
	// apiKey := os.Getenv("OPENAI_API_KEY_2")
	apiKey := "ollama"
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY_2 is not set")
	}
	config := openai.DefaultConfig(apiKey)
	// config.BaseURL = "https://vip.apiyi.com/v1"
	config.BaseURL = "http://192.168.1.111:11434/v1"
	client := openai.NewClientWithConfig(config)

	ctx := context.Background()

	return &OpenAI{
		Client: client,
		ctx:    ctx,
	}, nil
}

func (o *OpenAI) SendMessage(prompt, content string) (string, error) {
	req := openai.ChatCompletionRequest{
		// Model: openai.GPT4,
		Model: "qwen2.5",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: prompt,
			},
			{
				Role:    "User",
				Content: content,
			},
		},
	}

	resp, err := o.Client.CreateChatCompletion(o.ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}
	return resp.Choices[0].Message.Content, nil
}
