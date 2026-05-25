package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenRouterClient is an AI provider that uses the OpenRouter API.
// OpenRouter is a unified gateway for multiple AI providers (OpenAI, Anthropic, Google, etc.)
// using prefixed model names like "anthropic/claude-3-sonnet" or "openai/gpt-4".
type OpenRouterClient struct {
	client  *openai.Client
	model   string
	timeout time.Duration
}

// NewOpenRouterClient creates a new OpenRouter client with the given API key, model, and timeout.
// The model should be prefixed with the provider (e.g. "anthropic/claude-3-sonnet", "openai/gpt-4").
// The API key can also be set via the OPENROUTER_API_KEY environment variable.
func NewOpenRouterClient(apiKey, model string, timeout time.Duration) *OpenRouterClient {
	if model == "" {
		model = "openai/gpt-4"
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://openrouter.ai/api/v1"),
	)
	return &OpenRouterClient{
		client:  &client,
		model:   model,
		timeout: timeout,
	}
}

// Generate sends a prompt to OpenRouter and returns the generated text.
func (c *OpenRouterClient) Generate(ctx context.Context, prompt string) (string, error) {
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
		return "", fmt.Errorf("openrouter returned no completions. This is unexpected — try rephrasing your prompt or using a different model (current: %s)", c.model)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// enhanceError wraps OpenRouter API errors with actionable suggestions.
func (c *OpenRouterClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	// Timeout
	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") {
		return fmt.Errorf("openrouter request timed out after %s. Try increasing --timeout or reducing input size. Error: %w", c.timeout, err)
	}

	// Authentication errors
	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") ||
		strings.Contains(lower, "incorrect api key") || strings.Contains(lower, "invalid x-api-key") {
		return fmt.Errorf("openrouter authentication failed. Check your API key with:\n  goscribe provider add openrouter --key <YOUR_API_KEY> --model %s\nOr set the OPENROUTER_API_KEY environment variable.\nError: %w", c.model, err)
	}

	// Rate limiting
	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "429") ||
		strings.Contains(lower, "too many requests") {
		return fmt.Errorf("openrouter rate limit exceeded. Wait a moment and retry, or reduce --retries to fail fast. Error: %w", err)
	}

	// Quota/billing
	if strings.Contains(lower, "quota") || strings.Contains(lower, "billing") ||
		strings.Contains(lower, "insufficient") || strings.Contains(lower, "credits") {
		return fmt.Errorf("openrouter quota or credits exceeded. Check your account at https://openrouter.ai/credits. Error: %w", err)
	}

	// Model errors — common with provider-prefixed models
	if strings.Contains(lower, "model not found") || strings.Contains(lower, "invalid model") ||
		strings.Contains(lower, "no provider") || strings.Contains(lower, "provider not found") {
		return fmt.Errorf("openrouter model %q not found or provider unavailable. Check available models at https://openrouter.ai/models.\nUpdate with:\n  goscribe provider add openrouter --key <KEY> --model <PROVIDER/MODEL>\nError: %w", c.model, err)
	}

	// Context length
	if strings.Contains(lower, "context length") || strings.Contains(lower, "maximum context length") ||
		strings.Contains(lower, "token limit") {
		return fmt.Errorf("openrouter context length exceeded for model %s. Try reducing input size or use a model with larger context. Error: %w", c.model, err)
	}

	// Server errors
	if strings.Contains(lower, "500") || strings.Contains(lower, "502") ||
		strings.Contains(lower, "503") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("openrouter server error (temporary). The upstream provider may be experiencing issues. This usually resolves on retry. Error: %w", err)
	}

	// Network errors
	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "network") ||
		strings.Contains(lower, "i/o timeout") || strings.Contains(lower, "eof") {
		return fmt.Errorf("network error contacting OpenRouter. Check your internet connection. Error: %w", err)
	}

	return fmt.Errorf("openrouter error: %w", err)
}
