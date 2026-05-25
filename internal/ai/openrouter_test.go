package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewOpenRouterClient_DefaultModel(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "", 0)
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.model != "openai/gpt-4" {
		t.Fatalf("expected default model 'openai/gpt-4', got %q", client.model)
	}
	if client.client == nil {
		t.Fatal("expected OpenAI client to be initialized")
	}
}

func TestNewOpenRouterClient_CustomModel(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "anthropic/claude-3-sonnet", 0)
	if client.model != "anthropic/claude-3-sonnet" {
		t.Fatalf("expected model 'anthropic/claude-3-sonnet', got %q", client.model)
	}
}

func TestNewOpenRouterClient_WithTimeout(t *testing.T) {
	timeout := 30 * time.Second
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", timeout)
	if client.timeout != timeout {
		t.Fatalf("expected timeout %v, got %v", timeout, client.timeout)
	}
}

func TestOpenRouterClient_Generate_ContextTimeout(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 1*time.Millisecond)

	ctx := context.Background()
	_, err := client.Generate(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error due to timeout")
	}
	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("expected timeout-related error, got: %v", err)
	}
}

func TestOpenRouterClient_Generate_ContextCancellation(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Generate(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error due to canceled context")
	}
}

func TestOpenRouterClient_enhanceError_Nil(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)
	err := client.enhanceError(nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestOpenRouterClient_enhanceError_Timeout(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 30*time.Second)
	err := client.enhanceError(errors.New("request deadline exceeded"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestOpenRouterClient_enhanceError_Auth(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "anthropic/claude-3-sonnet", 0)

	tests := []struct {
		name    string
		errMsg  string
		wantSub string
	}{
		{"unauthorized", "unauthorized", "authentication failed"},
		{"invalid api key", "invalid api key provided", "authentication failed"},
		{"incorrect api key", "incorrect api key", "authentication failed"},
		{"invalid x-api-key", "invalid x-api-key header", "authentication failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("expected %q in error, got: %v", tt.wantSub, err)
			}
		})
	}
}

func TestOpenRouterClient_enhanceError_RateLimit(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"rate limit", "rate limit exceeded"},
		{"429", "429 too many requests"},
		{"too many requests", "too many requests"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "rate limit exceeded") {
				t.Fatalf("expected rate limit error, got: %v", err)
			}
		})
	}
}

func TestOpenRouterClient_enhanceError_Quota(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"insufficient", "insufficient credits"},
		{"credits", "credits exhausted"},
		{"quota", "quota exceeded"},
		{"billing", "billing issue"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "quota") && !strings.Contains(err.Error(), "credits") {
				t.Fatalf("expected quota/credits error, got: %v", err)
			}
		})
	}
}

func TestOpenRouterClient_enhanceError_ModelNotFound(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"model not found", "model not found"},
		{"invalid model", "invalid model"},
		{"no provider", "no provider available"},
		{"provider not found", "provider not found for model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), `model "openai/gpt-4" not found`) {
				t.Fatalf("expected model not found error, got: %v", err)
			}
		})
	}
}

func TestOpenRouterClient_enhanceError_ContextLength(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"context length", "maximum context length exceeded"},
		{"token limit", "token limit exceeded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "context length exceeded") {
				t.Fatalf("expected context length error, got: %v", err)
			}
		})
	}
}

func TestOpenRouterClient_enhanceError_ServerError(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"500", "500 internal server error"},
		{"502", "502 bad gateway"},
		{"503", "503 service unavailable"},
		{"internal server error", "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "server error") {
				t.Fatalf("expected server error, got: %v", err)
			}
		})
	}
}

func TestOpenRouterClient_enhanceError_NetworkError(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"connection refused", "connection refused"},
		{"network", "network is unreachable"},
		{"eof", "unexpected EOF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "network error") {
				t.Fatalf("expected network error, got: %v", err)
			}
		})
	}
}

func TestOpenRouterClient_enhanceError_Unknown(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)
	err := client.enhanceError(errors.New("some random error"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "openrouter error:") {
		t.Fatalf("expected generic openrouter error, got: %v", err)
	}
}

func TestOpenRouterClient_enhanceError_CaseInsensitive(t *testing.T) {
	client := NewOpenRouterClient("sk-or-test", "openai/gpt-4", 0)

	err := client.enhanceError(errors.New("TIMEOUT"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error for uppercase, got: %v", err)
	}
}

func TestOpenRouterClient_ProviderInterface(t *testing.T) {
	var _ Provider = (*OpenRouterClient)(nil)
}
