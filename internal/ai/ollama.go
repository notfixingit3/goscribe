// Package ai provides AI provider interfaces and implementations for text generation.
package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ollama/ollama/api"
)

// OllamaClient is an AI provider that uses a local Ollama instance.
type OllamaClient struct {
	client *api.Client
	model  string
}

// NewOllamaClient creates a new Ollama client for the given base URL and model.
func NewOllamaClient(baseURL, model string) (*OllamaClient, error) {
	if model == "" {
		model = "llama2"
	}

	var client *api.Client

	if baseURL != "" {
		u, parseErr := url.Parse(baseURL)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid Ollama URL %q: %w.\nMake sure the URL is in the format http://host:port (e.g. http://localhost:11434)", baseURL, parseErr)
		}
		if u.Scheme == "" {
			return nil, fmt.Errorf("invalid Ollama URL %q: missing scheme (use http://host:port)", baseURL)
		}
		client = api.NewClient(u, http.DefaultClient)
	} else {
		var err error
		client, err = api.ClientFromEnvironment()
		if err != nil {
			return nil, fmt.Errorf("could not create Ollama client from environment: %w.\nCheck that Ollama is running: ollama serve", err)
		}
	}

	return &OllamaClient{
		client: client,
		model:  model,
	}, nil
}

// Generate sends a prompt to Ollama and returns the generated text.
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
		return "", c.enhanceError(err)
	}

	return response, nil
}

// enhanceError wraps Ollama errors with actionable suggestions.
func (c *OllamaClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	// Connection refused - Ollama not running
	if strings.Contains(lower, "connection refused") {
		return fmt.Errorf("cannot connect to Ollama. Is it running?\n  Start it with: ollama serve\n  Or check your URL with: goscribe provider add ollama --url http://localhost:11434 --model %s\nError: %w", c.model, err)
	}

	// Model not found
	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no such model") ||
		strings.Contains(lower, "model '%s' not found") || strings.Contains(lower, "404") {
		return fmt.Errorf("ollama model %q not found. Pull it first:\n  ollama pull %s\nOr use a different model:\n  goscribe provider add ollama --model <MODEL>\nError: %w", c.model, c.model, err)
	}

	// Timeout
	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("ollama request timed out. The model %q may be loading for the first time or the prompt is too large.\nTry again or use a smaller model. Error: %w", c.model, err)
	}

	// Context cancellation
	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell // matching Ollama's error message
		return fmt.Errorf("ollama request was canceled. This usually means the timeout was too short. Try increasing --timeout. Error: %w", err)
	}

	// Server errors
	if strings.Contains(lower, "500") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("ollama server error (temporary). Check Ollama logs for details. Error: %w", err)
	}

	// Network errors
	if strings.Contains(lower, "network") || strings.Contains(lower, "eof") ||
		strings.Contains(lower, "connection reset") {
		return fmt.Errorf("network error contacting Ollama. Check that Ollama is running and reachable. Error: %w", err)
	}

	return fmt.Errorf("ollama error: %w", err)
}
