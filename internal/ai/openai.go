package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIClient is an AI provider that uses the OpenAI API.
type OpenAIClient struct {
	client  *openai.Client
	model   string
	timeout time.Duration
}

// NewOpenAIClient creates a new OpenAI client with the given API key, model, and timeout.
func NewOpenAIClient(apiKey, model string, timeout time.Duration) *OpenAIClient {
	if model == "" {
		model = "gpt-4"
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIClient{
		client:  &client,
		model:   model,
		timeout: timeout,
	}
}

// SetTimeout configures the request timeout for this client.
func (c *OpenAIClient) SetTimeout(d time.Duration) {
	c.timeout = d
}

// Generate sends a prompt to OpenAI and returns the generated text.
func (c *OpenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModel(c.model),
	})
	if err != nil {
		return "", c.enhanceError(err)
	}

	if len(chatCompletion.Choices) == 0 {
		return "", fmt.Errorf("openai returned no completions. This is unexpected — try rephrasing your prompt or using a different model (current: %s)", c.model)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// enhanceError wraps OpenAI API errors with actionable suggestions.
func (c *OpenAIClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	// Timeout
	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") {
		return fmt.Errorf("openai request timed out after %s. Try increasing --timeout or reducing input size. Error: %w", c.timeout, err)
	}

	// Authentication errors
	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") ||
		strings.Contains(lower, "incorrect api key") || strings.Contains(lower, "invalid x-api-key") {
		return fmt.Errorf("openai authentication failed. Check your API key with:\n  goscribe provider add openai --key <YOUR_API_KEY> --model %s\nError: %w", c.model, err)
	}

	// Rate limiting
	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "429") ||
		strings.Contains(lower, "too many requests") {
		return fmt.Errorf("openai rate limit exceeded. Wait a moment and retry, or reduce --retries to fail fast. Error: %w", err)
	}

	// Quota/billing
	if strings.Contains(lower, "quota") || strings.Contains(lower, "billing") ||
		strings.Contains(lower, "insufficient_quota") {
		return fmt.Errorf("openai quota exceeded. Check your billing details at https://platform.openai.com/account/billing. Error: %w", err)
	}

	// Model errors
	if strings.Contains(lower, "model not found") || strings.Contains(lower, "invalid model") {
		return fmt.Errorf("openai model %q not found. Check available models at https://platform.openai.com/docs/models. Update with:\n  goscribe provider add openai --key <KEY> --model <MODEL>\nError: %w", c.model, err)
	}

	// Context length
	if strings.Contains(lower, "context length") || strings.Contains(lower, "maximum context length") {
		return fmt.Errorf("openai context length exceeded for model %s. Try reducing input size or use a model with larger context (e.g. gpt-4-turbo). Error: %w", c.model, err)
	}

	// Server errors
	if strings.Contains(lower, "500") || strings.Contains(lower, "502") ||
		strings.Contains(lower, "503") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("openai server error (temporary). This usually resolves on retry. Error: %w", err)
	}

	// Network errors
	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "network") ||
		strings.Contains(lower, "i/o timeout") || strings.Contains(lower, "eof") {
		return fmt.Errorf("network error contacting OpenAI. Check your internet connection. Error: %w", err)
	}

	return fmt.Errorf("openai error: %w", err)
}
