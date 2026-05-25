package ai

import (
	"context"
	"fmt"
	"strings"
)

// GitHubCopilotClient is an AI provider that uses GitHub Copilot.
type GitHubCopilotClient struct {
	openAI *OpenAIClient
	url    string
}

// NewGitHubCopilotClient creates a new GitHub Copilot client with the given token, URL, and model.
func NewGitHubCopilotClient(token, url, model string) *GitHubCopilotClient {
	if url == "" {
		url = "https://api.githubcopilot.com"
	}
	if model == "" {
		model = "gpt-4"
	}

	client := NewOpenAIClient(token, model, 0)
	return &GitHubCopilotClient{
		openAI: client,
		url:    url,
	}
}

// Generate sends a prompt to GitHub Copilot and returns the generated text.
func (c *GitHubCopilotClient) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.openAI.Generate(ctx, prompt)
	if err != nil {
		return "", c.enhanceError(err)
	}
	return result, nil
}

func (c *GitHubCopilotClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "connection reset") {
		return fmt.Errorf("cannot connect to GitHub Copilot at %s. Check your internet connection.\nError: %w", c.url, err)
	}

	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid token") ||
		strings.Contains(lower, "authentication") || strings.Contains(lower, "401") {
		return fmt.Errorf("github-copilot authentication failed. Check your GitHub token and Copilot subscription.\nError: %w", err)
	}

	if strings.Contains(lower, "model not found") || strings.Contains(lower, "no such model") {
		return fmt.Errorf("github-copilot model %q not found. Available models depend on your Copilot subscription.\nError: %w", c.openAI.model, err)
	}

	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("github-copilot request timed out. The model may be loading or the prompt is too large.\nError: %w", err)
	}

	if strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell
		return fmt.Errorf("github-copilot request was canceled. Error: %w", err)
	}

	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "too many requests") ||
		strings.Contains(lower, "quota exceeded") {
		return fmt.Errorf("github-copilot rate limit exceeded. Check your Copilot quota.\nError: %w", err)
	}

	if strings.Contains(lower, "403") || strings.Contains(lower, "forbidden") {
		return fmt.Errorf("github-copilot access denied. Ensure you have an active Copilot subscription.\nError: %w", err)
	}

	if strings.Contains(lower, "500") || strings.Contains(lower, "internal server error") {
		return fmt.Errorf("github-copilot server error. Check GitHub status for outages.\nError: %w", err)
	}

	if strings.Contains(lower, "network") || strings.Contains(lower, "eof") {
		return fmt.Errorf("network error contacting GitHub Copilot. Check your connection.\nError: %w", err)
	}

	return fmt.Errorf("github-copilot error: %w", err)
}
