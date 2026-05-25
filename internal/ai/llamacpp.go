package ai

import (
	"context"
	"fmt"
	"strings"
)

// LLaMACppClient is an AI provider that uses a local LLaMA.cpp server.
type LLaMACppClient struct {
	openAI *OpenAIClient
	url    string
}

// NewLLaMACppClient creates a new LLaMA.cpp client with the given URL and model.
func NewLLaMACppClient(url, model string) *LLaMACppClient {
	if url == "" {
		url = "http://localhost:8080"
	}
	if model == "" {
		model = "llama2"
	}

	client := NewOpenAIClient("not-needed", model, 0)
	return &LLaMACppClient{
		openAI: client,
		url:    url,
	}
}

// Generate sends a prompt to LLaMA.cpp and returns the generated text.
func (c *LLaMACppClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.openAI.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *LLaMACppClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") {
		return fmt.Errorf("cannot connect to LLaMA.cpp server at %s. Is it running?\n  Start the server with: ./server -m model.gguf\n  Or check your URL with: goscribe provider add llamacpp --url %s --model %s\nError: %w", c.url, c.url, c.openAI.model, err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no model") {
		return fmt.Errorf("llamacpp model not found. Make sure the model file exists and is loaded.\nError: %w", err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("llamacpp request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("llamacpp request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("llamacpp server error. Check server logs. Error: %w", err)
	}

	if strings.Contains(lower, "network") || strings.Contains(lower, "eof") ||
		strings.Contains(lower, "connection reset") {
		return fmt.Errorf("network error contacting LLaMA.cpp at %s. Check that the server is running. Error: %w", c.url, err)
	}

	return fmt.Errorf("llamacpp error: %w", err)
}
