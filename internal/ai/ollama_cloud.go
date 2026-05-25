package ai

import (
	"context"
	"fmt"
	"strings"
)

// OllamaCloudClient is an AI provider that uses a managed Ollama Cloud service.
type OllamaCloudClient struct {
	ollama *OllamaClient
	apiKey string
}

// NewOllamaCloudClient creates a new Ollama Cloud client with the given API key, URL, and model.
func NewOllamaCloudClient(apiKey, url, model string) (*OllamaCloudClient, error) {
	if url == "" {
		url = "https://cloud.ollama.com"
	}
	if model == "" {
		model = "llama2"
	}

	ollama, err := NewOllamaClient(url, model)
	if err != nil {
		return nil, err
	}

	return &OllamaCloudClient{
		ollama: ollama,
		apiKey: apiKey,
	}, nil
}

// Generate sends a prompt to Ollama Cloud and returns the generated text.
func (c *OllamaCloudClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.ollama.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *OllamaCloudClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "connection reset") {
		return fmt.Errorf("cannot connect to Ollama Cloud. Check your internet connection and URL.\nError: %w", err)
	}

	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") ||
		strings.Contains(lower, "authentication") {
		return fmt.Errorf("ollama-cloud authentication failed. Check your API key.\nError: %w", err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no such model") {
		return fmt.Errorf("ollama-cloud model not found. Check available models in your Ollama Cloud dashboard.\nError: %w", err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("ollama-cloud request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("ollama-cloud request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "502") || strings.Contains(lower, "503") {
		return fmt.Errorf("ollama-cloud server error (temporary). This usually resolves on retry. Error: %w", err)
	}

	return fmt.Errorf("ollama-cloud error: %w", err)
}
