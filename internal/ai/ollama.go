package ai

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ollama/ollama/api"
)

type OllamaClient struct {
	client *api.Client
	model  string
}

func NewOllamaClient(baseURL, model string) (*OllamaClient, error) {
	if model == "" {
		model = "llama2"
	}

	var client *api.Client
	var err error

	if baseURL != "" {
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("parse url: %w", err)
		}
		client = api.NewClient(u, nil)
	} else {
		client, err = api.ClientFromEnvironment()
		if err != nil {
			return nil, fmt.Errorf("create client: %w", err)
		}
	}

	return &OllamaClient{
		client: client,
		model:  model,
	}, nil
}

func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	var response string
	err := c.client.Generate(ctx, &api.GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
	}, func(gr api.GenerateResponse) error {
		response += gr.Response
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("generate: %w", err)
	}

	return response, nil
}
