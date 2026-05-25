package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// GeminiClient is an AI provider that uses the Google Gemini API.
type GeminiClient struct {
	apiKey  string
	model   string
	timeout time.Duration
	baseURL string
	client  *http.Client
}

// NewGeminiClient creates a new Gemini client with the given API key, model, and timeout.
// If apiKey is empty, it falls back to the GOOGLE_API_KEY environment variable.
func NewGeminiClient(apiKey, model string, timeout time.Duration) *GeminiClient {
	if model == "" {
		model = "gemini-2.0-flash"
	}
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	return &GeminiClient{
		apiKey:  apiKey,
		model:   model,
		timeout: timeout,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		client:  &http.Client{},
	}
}

// SetTimeout configures the request timeout for this client.
func (c *GeminiClient) SetTimeout(d time.Duration) {
	c.timeout = d
}

// geminiRequest represents the request body for the Gemini generateContent API.
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

// geminiContent represents a content message in the Gemini API.
type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// geminiPart represents a text part within a content message.
type geminiPart struct {
	Text string `json:"text"`
}

// geminiResponse represents the response from the Gemini API.
type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *geminiError      `json:"error,omitempty"`
}

// geminiCandidate represents a candidate response from Gemini.
type geminiCandidate struct {
	Content       geminiContent        `json:"content"`
	FinishReason  string               `json:"finishReason"`
	SafetyRatings []geminiSafetyRating `json:"safetyRatings"`
}

// geminiSafetyRating represents safety assessment for a response.
type geminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// geminiError represents an error from the Gemini API.
type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Generate sends a prompt to Gemini and returns the generated text.
func (c *GeminiClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("gemini API key is required.\nSet it with:\n  goscribe provider add gemini --key <YOUR_API_KEY> --model %s\nOr set the GOOGLE_API_KEY environment variable.\nGet an API key at https://aistudio.google.com/apikey", c.model)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{{Text: prompt}},
				Role:  "user",
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gemini: failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("gemini: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", c.enhanceError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gemini: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", c.enhanceHTTPError(resp.StatusCode, respBytes)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		return "", fmt.Errorf("gemini: failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		return "", c.enhanceError(fmt.Errorf("%s: %s", geminiResp.Error.Status, geminiResp.Error.Message))
	}

	if len(geminiResp.Candidates) == 0 {
		return "", fmt.Errorf("gemini returned no candidates. The prompt may have been blocked by safety filters.\nTry rephrasing your prompt or using a different model (current: %s)", c.model)
	}

	candidate := geminiResp.Candidates[0]

	if candidate.FinishReason == "SAFETY" {
		return "", fmt.Errorf("gemini blocked the response due to safety filters.\nThis means the generated content was flagged as potentially unsafe.\nTry rephrasing your prompt or using a different model (current: %s)", c.model)
	}

	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty content (finishReason: %s, model: %s).\nTry rephrasing your prompt", candidate.FinishReason, c.model)
	}

	return candidate.Content.Parts[0].Text, nil
}

// enhanceError wraps Gemini API errors with actionable suggestions.
func (c *GeminiClient) enhanceError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	// Timeout
	if strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "context canceled") || strings.Contains(lower, "context cancelled") { //nolint:misspell // matching error messages
		return fmt.Errorf("gemini request timed out after %s. Try increasing --timeout or reducing input size. Error: %w", c.timeout, err)
	}

	// Authentication errors
	if strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") ||
		strings.Contains(lower, "api key not valid") || strings.Contains(lower, "401") ||
		strings.Contains(lower, "permission_denied") {
		return fmt.Errorf("gemini authentication failed. Check your API key with:\n  goscribe provider add gemini --key <YOUR_API_KEY> --model %s\nOr set GOOGLE_API_KEY environment variable.\nGet an API key at https://aistudio.google.com/apikey\nError: %w", c.model, err)
	}

	// Rate limiting
	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "429") ||
		strings.Contains(lower, "too many requests") || strings.Contains(lower, "resource_exhausted") {
		return fmt.Errorf("gemini rate limit exceeded. Wait a moment and retry, or reduce --retries to fail fast. Error: %w", err)
	}

	// Quota/billing
	if strings.Contains(lower, "quota") || strings.Contains(lower, "billing") {
		return fmt.Errorf("gemini quota exceeded. Check your billing details at https://console.cloud.google.com/billing. Error: %w", err)
	}

	// Model errors
	if strings.Contains(lower, "model not found") || strings.Contains(lower, "invalid model") ||
		strings.Contains(lower, "not_found") || strings.Contains(lower, "does not exist") {
		return fmt.Errorf("gemini model %q not found. Check available models at https://ai.google.dev/models.\nUpdate with:\n  goscribe provider add gemini --key <KEY> --model <MODEL>\nError: %w", c.model, err)
	}

	// Context length / token limit
	if strings.Contains(lower, "context length") || strings.Contains(lower, "token limit") ||
		strings.Contains(lower, "max tokens") {
		return fmt.Errorf("gemini context length exceeded for model %s. Try reducing input size or use a model with larger context (e.g. gemini-2.0-flash). Error: %w", c.model, err)
	}

	// Safety blocks
	if strings.Contains(lower, "safety") || strings.Contains(lower, "blocked") {
		return fmt.Errorf("gemini response blocked by safety filters. Try rephrasing your prompt. Error: %w", err)
	}

	// Server errors
	if strings.Contains(lower, "500") || strings.Contains(lower, "502") ||
		strings.Contains(lower, "503") || strings.Contains(lower, "internal server error") ||
		strings.Contains(lower, "internal_error") {
		return fmt.Errorf("gemini server error (temporary). This usually resolves on retry. Error: %w", err)
	}

	// Network errors
	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "network") ||
		strings.Contains(lower, "i/o timeout") || strings.Contains(lower, "eof") ||
		strings.Contains(lower, "dns") {
		return fmt.Errorf("network error contacting Gemini. Check your internet connection. Error: %w", err)
	}

	return fmt.Errorf("gemini error: %w", err)
}

// enhanceHTTPError wraps non-200 HTTP responses with actionable suggestions.
func (c *GeminiClient) enhanceHTTPError(statusCode int, body []byte) error {
	var errResp struct {
		Error *geminiError `json:"error"`
	}
	apiMsg := ""
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
		apiMsg = errResp.Error.Message
	}

	switch {
	case statusCode == 400:
		return c.enhanceError(fmt.Errorf("bad request: %s", apiMsg))
	case statusCode == 401 || statusCode == 403:
		return c.enhanceError(fmt.Errorf("unauthorized: %s", apiMsg))
	case statusCode == 404:
		return c.enhanceError(fmt.Errorf("model not found: %s", apiMsg))
	case statusCode == 429:
		return c.enhanceError(fmt.Errorf("rate limit: %s", apiMsg))
	case statusCode >= 500:
		return c.enhanceError(fmt.Errorf("server error (%d): %s", statusCode, apiMsg))
	default:
		return fmt.Errorf("gemini returned HTTP %d: %s", statusCode, string(body))
	}
}
