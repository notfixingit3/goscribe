package ai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIClient struct {
	client *openai.Client
	model  string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4"
	}
	
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIClient{
		client: &client,
		model:  model,
	}
}

func (c *OpenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModel(c.model),
	})
	if err != nil {
		return "", fmt.Errorf("chat completion: %w", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
