package ai

import (
	"context"
	"fmt"
	"strings"
)

// OllamaSelfHostedClient is an AI provider for custom Ollama deployments.
type OllamaSelfHostedClient struct {
	ollama *OllamaClient
}

// NewOllamaSelfHostedClient creates a new self-hosted Ollama client.
func NewOllamaSelfHostedClient(url, model string) (*OllamaSelfHostedClient, error) {
	if url == "" {
		return nil, fmt.Errorf("self-hosted Ollama requires a URL; set it with: goscribe provider add ollama-selfhosted --url http://your-server:11434 --model llama2")
	}
	if model == "" {
		model = "llama2"
	}

	ollama, err := NewOllamaClient(url, model)
	if err != nil {
		return nil, err
	}

	return &OllamaSelfHostedClient{
		ollama: ollama,
	}, nil
}

// Generate sends a prompt to self-hosted Ollama and returns the generated text.
func (c *OllamaSelfHostedClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.ollama.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *OllamaSelfHostedClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "connection reset") ||
		strings.Contains(lower, "no such host") {
		return fmt.Errorf("cannot connect to self-hosted Ollama. Check that the server is running and accessible.\nError: %w", err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no such model") {
		return fmt.Errorf("self-hosted ollama model not found. Pull it with: ollama pull %s\nError: %w", c.ollama.model, err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("self-hosted ollama request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("self-hosted ollama request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "502") || strings.Contains(lower, "503") {
		return fmt.Errorf("self-hosted ollama server error. Check server logs. Error: %w", err)
	}

	return fmt.Errorf("self-hosted ollama error: %w", err)
}
