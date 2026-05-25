package ai

import (
	"context"
	"fmt"
	"strings"
)

// OpenCodeClient is an AI provider that connects to the OpenCode agent.
type OpenCodeClient struct {
	openAI *OpenAIClient
	url    string
}

// NewOpenCodeClient creates a new OpenCode client with the given URL and model.
func NewOpenCodeClient(url, model string) *OpenCodeClient {
	if url == "" {
		url = "http://localhost:8080"
	}
	if model == "" {
		model = "default"
	}

	client := NewOpenAIClient("not-needed", model, 0)
	return &OpenCodeClient{
		openAI: client,
		url:    url,
	}
}

// Generate sends a prompt to OpenCode and returns the generated text.
func (c *OpenCodeClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.openAI.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *OpenCodeClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") {
		return fmt.Errorf("cannot connect to OpenCode agent at %s. Is it running?\n  Start the agent or check your URL with: goscribe provider add opencode --url %s --model %s\nError: %w", c.url, c.url, c.openAI.model, err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no model") {
		return fmt.Errorf("opencode model not found. Check agent configuration.\nError: %w", err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("opencode request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("opencode request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("opencode agent error. Check agent logs. Error: %w", err)
	}

	if strings.Contains(lower, "network") || strings.Contains(lower, "eof") ||
		strings.Contains(lower, "connection reset") {
		return fmt.Errorf("network error contacting OpenCode agent at %s. Check that the agent is running. Error: %w", c.url, err)
	}

	return fmt.Errorf("opencode error: %w", err)
}
