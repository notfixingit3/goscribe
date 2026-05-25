package ai

import (
	"context"
	"fmt"
	"strings"
)

// VLLMClient is an AI provider that uses a local vLLM server.
type VLLMClient struct {
	openAI *OpenAIClient
	url    string
}

// NewVLLMClient creates a new vLLM client with the given URL and model.
func NewVLLMClient(url, model string) *VLLMClient {
	if url == "" {
		url = "http://localhost:8000"
	}
	if model == "" {
		model = "default"
	}

	client := NewOpenAIClient("not-needed", model, 0)
	return &VLLMClient{
		openAI: client,
		url:    url,
	}
}

// Generate sends a prompt to vLLM and returns the generated text.
func (c *VLLMClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.openAI.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *VLLMClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") {
		return fmt.Errorf("cannot connect to vLLM at %s. Is it running?\n  Start vLLM with: vllm serve <model>\n  Or check your URL with: goscribe provider add vllm --url %s --model %s\nError: %w", c.url, c.url, c.openAI.model, err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no model") ||
		strings.Contains(lower, "model not loaded") {
		return fmt.Errorf("vllm model not loaded. Start vLLM with the model: vllm serve %s\nError: %w", c.openAI.model, err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("vllm request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("vllm request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("vllm server error. Check vLLM logs.\nError: %w", err)
	}

	if strings.Contains(lower, "network") || strings.Contains(lower, "eof") ||
		strings.Contains(lower, "connection reset") {
		return fmt.Errorf("network error contacting vLLM at %s. Check that vLLM is running.\nError: %w", c.url, err)
	}

	if strings.Contains(lower, "cuda") || strings.Contains(lower, "gpu") ||
		strings.Contains(lower, "out of memory") {
		return fmt.Errorf("vllm GPU error. Check GPU availability and memory.\nError: %w", err)
	}

	return fmt.Errorf("vllm error: %w", err)
}
