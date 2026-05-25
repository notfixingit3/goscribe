package ai

import (
	"context"
	"fmt"
	"strings"
)

// ZaiClient is an AI provider that uses Z.ai.
type ZaiClient struct {
	openAI *OpenAIClient
	url    string
}

// NewZaiClient creates a new Z.ai client with the given API key, URL, and model.
func NewZaiClient(apiKey, url, model string) *ZaiClient {
	if url == "" {
		url = "https://api.z.ai"
	}
	if model == "" {
		model = "z-large"
	}

	client := NewOpenAIClient(apiKey, model, 0)
	return &ZaiClient{
		openAI: client,
		url:    url,
	}
}

// Generate sends a prompt to Z.ai and returns the generated text.
func (c *ZaiClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.openAI.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *ZaiClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "connection reset") {
		return fmt.Errorf("cannot connect to Z.ai at %s. Check your internet connection and URL.\nError: %w", c.url, err)
	}

	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") ||
		strings.Contains(lower, "authentication") {
		return fmt.Errorf("zai authentication failed. Check your API key at https://z.ai.\nError: %w", err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no such model") {
		return fmt.Errorf("zai model %q not found. Check available models at https://z.ai.\nError: %w", c.openAI.model, err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("zai request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("zai request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "too many requests") {
		return fmt.Errorf("zai rate limit exceeded. Wait a moment and retry.\nError: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("zai server error. Check https://status.z.ai for outages.\nError: %w", err)
	}

	if strings.Contains(lower, "network") || strings.Contains(lower, "eof") {
		return fmt.Errorf("network error contacting Z.ai. Check your connection.\nError: %w", err)
	}

	return fmt.Errorf("zai error: %w", err)
}
