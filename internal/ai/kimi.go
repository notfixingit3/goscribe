package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	// DefaultKimiBaseURL is the default Moonshot AI API endpoint.
	DefaultKimiBaseURL = "https://api.moonshot.cn/v1"
	// DefaultKimiModel is the default Moonshot model.
	DefaultKimiModel = "moonshot-v1-8k"
)

// KimiClient is an AI provider that uses the Moonshot AI (Kimi) API.
type KimiClient struct {
	client  *openai.Client
	model   string
	timeout time.Duration
}

// NewKimiClient creates a new Kimi client with the given API key, model, and timeout.
// The Moonshot API is OpenAI-compatible, so it uses the openai-go library with a custom base URL.
func NewKimiClient(apiKey, model string, timeout time.Duration) *KimiClient {
	if model == "" {
		model = DefaultKimiModel
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(DefaultKimiBaseURL),
	)
	return &KimiClient{
		client:  &client,
		model:   model,
		timeout: timeout,
	}
}

// Generate sends a prompt to Moonshot AI and returns the generated text.
func (c *KimiClient) Generate(ctx context.Context, prompt string) (string, error) {
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
		return "", fmt.Errorf("kimi returned no completions. This is unexpected — try rephrasing your prompt or using a different model (current: %s)", c.model)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// enhanceError wraps Kimi API errors with actionable suggestions.
func (c *KimiClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	// Timeout
	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") {
		return fmt.Errorf("kimi request timed out after %s. Try increasing --timeout or reducing input size. Error: %w", c.timeout, err)
	}

	// Authentication errors
	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") ||
		strings.Contains(lower, "incorrect api key") || strings.Contains(lower, "authentication") {
		return fmt.Errorf("kimi authentication failed. Check your API key with:\n  goscribe provider add kimi --key <YOUR_API_KEY> --model %s\nYou can get a key at https://platform.moonshot.cn\nError: %w", c.model, err)
	}

	// Rate limiting
	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "429") ||
		strings.Contains(lower, "too many requests") {
		return fmt.Errorf("kimi rate limit exceeded. Wait a moment and retry, or reduce --retries to fail fast. Error: %w", err)
	}

	// Quota/billing
	if strings.Contains(lower, "quota") || strings.Contains(lower, "billing") ||
		strings.Contains(lower, "insufficient") {
		return fmt.Errorf("kimi quota exceeded. Check your billing details at https://platform.moonshot.cn. Error: %w", err)
	}

	// Model errors
	if strings.Contains(lower, "model not found") || strings.Contains(lower, "invalid model") {
		return fmt.Errorf("kimi model %q not found. Available models: moonshot-v1-8k, moonshot-v1-32k, moonshot-v1-128k.\nUpdate with:\n  goscribe provider add kimi --key <KEY> --model <MODEL>\nError: %w", c.model, err)
	}

	// Context length
	if strings.Contains(lower, "context length") || strings.Contains(lower, "maximum context length") ||
		strings.Contains(lower, "token limit") {
		return fmt.Errorf("kimi context length exceeded for model %s. Try reducing input size or use a model with larger context (e.g. moonshot-v1-128k). Error: %w", c.model, err)
	}

	// Server errors
	if strings.Contains(lower, "500") || strings.Contains(lower, "502") ||
		strings.Contains(lower, "503") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("kimi server error (temporary). This usually resolves on retry. Error: %w", err)
	}

	// Network errors
	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "network") ||
		strings.Contains(lower, "i/o timeout") || strings.Contains(lower, "eof") {
		return fmt.Errorf("network error contacting Kimi API. Check your internet connection. Error: %w", err)
	}

	return fmt.Errorf("kimi error: %w", err)
}
