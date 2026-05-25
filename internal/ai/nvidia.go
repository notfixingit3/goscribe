package ai

import (
	"context"
	"fmt"
	"strings"
)

// NVIDIAClient is an AI provider that uses the NVIDIA NIM API.
type NVIDIAClient struct {
	openAI *OpenAIClient
	url    string
}

// NewNVIDIAClient creates a new NVIDIA client with the given API key, URL, and model.
func NewNVIDIAClient(apiKey, url, model string) *NVIDIAClient {
	if url == "" {
		url = "https://integrate.api.nvidia.com"
	}
	if model == "" {
		model = "meta/llama-3.1-405b-instruct"
	}

	client := NewOpenAIClient(apiKey, model, 0)
	return &NVIDIAClient{
		openAI: client,
		url:    url,
	}
}

// Generate sends a prompt to NVIDIA NIM and returns the generated text.
func (c *NVIDIAClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.openAI.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *NVIDIAClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "connection reset") {
		return fmt.Errorf("cannot connect to NVIDIA NIM at %s. Check your internet connection and URL.\nError: %w", c.url, err)
	}

	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") ||
		strings.Contains(lower, "authentication") {
		return fmt.Errorf("nvidia authentication failed. Check your API key at https://build.nvidia.com.\nError: %w", err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no such model") {
		return fmt.Errorf("nvidia model %q not found. Browse available models at https://build.nvidia.com.\nError: %w", c.openAI.model, err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("nvidia request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("nvidia request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "too many requests") {
		return fmt.Errorf("nvidia rate limit exceeded. Check your quota at https://build.nvidia.com.\nError: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("nvidia server error. Check NVIDIA status page for outages.\nError: %w", err)
	}

	if strings.Contains(lower, "network") || strings.Contains(lower, "eof") {
		return fmt.Errorf("network error contacting NVIDIA NIM. Check your connection.\nError: %w", err)
	}

	return fmt.Errorf("nvidia error: %w", err)
}
