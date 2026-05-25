package ai

import (
	"context"
	"fmt"
	"strings"
)

// LMStudioClient is an AI provider that uses LM Studio's local OpenAI-compatible API.
type LMStudioClient struct {
	openAI *OpenAIClient
	url    string
}

// NewLMStudioClient creates a new LM Studio client with the given URL and model.
func NewLMStudioClient(url, model string) *LMStudioClient {
	if url == "" {
		url = "http://localhost:1234"
	}
	if model == "" {
		model = "local-model"
	}

	client := NewOpenAIClient("not-needed", model, 0)
	return &LMStudioClient{
		openAI: client,
		url:    url,
	}
}

// Generate sends a prompt to LM Studio and returns the generated text.
func (c *LMStudioClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.openAI.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *LMStudioClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") {
		return fmt.Errorf("cannot connect to LM Studio at %s. Is it running?\n  Start LM Studio and load a model\n  Or check your URL with: goscribe provider add lmstudio --url %s --model %s\nError: %w", c.url, c.url, c.openAI.model, err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no model") ||
		strings.Contains(lower, "model not loaded") {
		return fmt.Errorf("lmstudio model not loaded. Open LM Studio and load a model first.\nError: %w", err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("lmstudio request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("lmstudio request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("lmstudio server error. Check LM Studio logs. Error: %w", err)
	}

	if strings.Contains(lower, "network") || strings.Contains(lower, "eof") ||
		strings.Contains(lower, "connection reset") {
		return fmt.Errorf("network error contacting LM Studio at %s. Check that LM Studio is running. Error: %w", c.url, err)
	}

	return fmt.Errorf("lmstudio error: %w", err)
}
