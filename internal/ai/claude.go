package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeClient is an AI provider that uses the Anthropic Claude Messages API.
type ClaudeClient struct {
	client  *anthropic.Client
	model   string
	timeout time.Duration
}

// NewClaudeClient creates a new Claude client with the given API key, model, and timeout.
func NewClaudeClient(apiKey, model string, timeout time.Duration) *ClaudeClient {
	if model == "" {
		model = string(anthropic.ModelClaudeSonnet4_6)
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &ClaudeClient{
		client:  &client,
		model:   model,
		timeout: timeout,
	}
}

// SetTimeout configures the request timeout for this client.
func (c *ClaudeClient) SetTimeout(d time.Duration) {
	c.timeout = d
}

// Generate sends a prompt to the Anthropic Claude Messages API and returns the generated text.
func (c *ClaudeClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Model: c.model,
	})
	if err != nil {
		return "", c.enhanceError(err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("claude returned no content. This is unexpected — try rephrasing your prompt or using a different model (current: %s)", c.model)
	}

	// Extract text from response content blocks
	var sb strings.Builder
	for _, block := range message.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}

	if sb.Len() == 0 {
		return "", fmt.Errorf("claude returned no text content. This is unexpected — try rephrasing your prompt or using a different model (current: %s)", c.model)
	}

	return sb.String(), nil
}

// enhanceError wraps Claude API errors with actionable suggestions.
func (c *ClaudeClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	// Timeout
	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") {
		return fmt.Errorf("claude request timed out after %s. Try increasing --timeout or reducing input size. Error: %w", c.timeout, err)
	}

	// Authentication errors
	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") ||
		strings.Contains(lower, "invalid x-api-key") || strings.Contains(lower, "authentication_error") {
		return fmt.Errorf("claude authentication failed. Check your API key with:\n  goscribe provider add claude --key <YOUR_API_KEY> --model %s\nError: %w", c.model, err)
	}

	// Rate limiting
	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "429") ||
		strings.Contains(lower, "too many requests") {
		return fmt.Errorf("claude rate limit exceeded. Wait a moment and retry, or reduce --retries to fail fast. Error: %w", err)
	}

	// Quota/billing
	if strings.Contains(lower, "quota") || strings.Contains(lower, "billing") ||
		strings.Contains(lower, "insufficient") {
		return fmt.Errorf("claude quota exceeded. Check your billing details at https://console.anthropic.com/settings/billing. Error: %w", err)
	}

	// Model errors
	if strings.Contains(lower, "model not found") || strings.Contains(lower, "invalid model") ||
		strings.Contains(lower, "not_found_error") {
		return fmt.Errorf("claude model %q not found. Check available models at https://docs.anthropic.com/en/docs/about-claude/models. Update with:\n  goscribe provider add claude --key <KEY> --model <MODEL>\nError: %w", c.model, err)
	}

	// Context length
	if strings.Contains(lower, "context length") || strings.Contains(lower, "context window") ||
		strings.Contains(lower, "max tokens") || strings.Contains(lower, "prompt is too long") {
		return fmt.Errorf("claude context length exceeded for model %s. Try reducing input size or use a model with larger context (e.g. claude-3-5-sonnet-latest). Error: %w", c.model, err)
	}

	// Server errors
	if strings.Contains(lower, "500") || strings.Contains(lower, "502") ||
		strings.Contains(lower, "503") || strings.Contains(lower, "internal server error") ||
		strings.Contains(lower, "api_error") || strings.Contains(lower, "overloaded_error") {
		return fmt.Errorf("claude server error (temporary). This usually resolves on retry. Error: %w", err)
	}

	// Network errors
	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "network is unreachable") ||
		strings.Contains(lower, "eof") || strings.Contains(lower, "unexpected eof") {
		return fmt.Errorf("network error contacting Anthropic. Check your internet connection. Error: %w", err)
	}

	return fmt.Errorf("claude error: %w", err)
}
