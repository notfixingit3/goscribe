package ai

import (
	"context"
	"fmt"
	"strings"
)

// XAIClient is an AI provider that uses the xAI Grok API.
type XAIClient struct {
	openAI *OpenAIClient
	url    string
}

// NewXAIClient creates a new xAI client with the given API key, URL, and model.
func NewXAIClient(apiKey, url, model string) *XAIClient {
	if url == "" {
		url = "https://api.x.ai"
	}
	if model == "" {
		model = "grok-2"
	}

	client := NewOpenAIClient(apiKey, model, 0)
	return &XAIClient{
		openAI: client,
		url:    url,
	}
}

// Generate sends a prompt to xAI Grok and returns the generated text.
func (c *XAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.openAI.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *XAIClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "connection reset") {
		return fmt.Errorf("cannot connect to xAI at %s. Check your internet connection and URL.\nError: %w", c.url, err)
	}

	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") ||
		strings.Contains(lower, "authentication") {
		return fmt.Errorf("xai authentication failed. Check your API key at https://console.x.ai.\nError: %w", err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no such model") {
		return fmt.Errorf("xai model %q not found. Available models: grok-2, grok-2-mini.\nError: %w", c.openAI.model, err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("xai request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("xai request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "too many requests") {
		return fmt.Errorf("xai rate limit exceeded. Wait a moment and retry.\nError: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("xai server error. Check https://status.x.ai for outages.\nError: %w", err)
	}

	if strings.Contains(lower, "network") || strings.Contains(lower, "eof") {
		return fmt.Errorf("network error contacting xAI. Check your connection.\nError: %w", err)
	}

	return fmt.Errorf("xai error: %w", err)
}
