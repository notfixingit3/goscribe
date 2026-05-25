package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewClaudeClient_DefaultModel(t *testing.T) {
	client := NewClaudeClient("sk-test", "", 0)
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.model != "claude-sonnet-4-6" {
		t.Fatalf("expected default model 'claude-sonnet-4-6', got %q", client.model)
	}
	if client.client == nil {
		t.Fatal("expected Claude client to be initialized")
	}
}

func TestNewClaudeClient_CustomModel(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-opus-4-6", 0)
	if client.model != "claude-opus-4-6" {
		t.Fatalf("expected model 'claude-opus-4-6', got %q", client.model)
	}
}

func TestNewClaudeClient_WithTimeout(t *testing.T) {
	timeout := 30 * time.Second
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", timeout)
	if client.timeout != timeout {
		t.Fatalf("expected timeout %v, got %v", timeout, client.timeout)
	}
}

func TestClaudeClient_SetTimeout(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)
	newTimeout := 2 * time.Minute
	client.SetTimeout(newTimeout)
	if client.timeout != newTimeout {
		t.Fatalf("expected timeout %v after SetTimeout, got %v", newTimeout, client.timeout)
	}
}

func TestClaudeClient_Generate_ContextTimeout(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 1*time.Millisecond)

	ctx := context.Background()
	_, err := client.Generate(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error due to timeout")
	}
	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("expected timeout-related error, got: %v", err)
	}
}

func TestClaudeClient_Generate_ContextCancellation(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Generate(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error due to canceled context")
	}
}

func TestClaudeClient_enhanceError_Nil(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)
	err := client.enhanceError(nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestClaudeClient_enhanceError_Timeout(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 30*time.Second)
	err := client.enhanceError(errors.New("request deadline exceeded"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestClaudeClient_enhanceError_Auth(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)

	tests := []struct {
		name    string
		errMsg  string
		wantSub string
	}{
		{"unauthorized", "unauthorized", "authentication failed"},
		{"invalid api key", "invalid api key provided", "authentication failed"},
		{"invalid x-api-key", "invalid x-api-key header", "authentication failed"},
		{"authentication_error", "authentication_error: invalid", "authentication failed"},
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

func TestClaudeClient_enhanceError_RateLimit(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)

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

func TestClaudeClient_enhanceError_Quota(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"quota", "quota exceeded"},
		{"billing", "billing issue"},
		{"insufficient", "insufficient credits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "quota exceeded") {
				t.Fatalf("expected quota error, got: %v", err)
			}
		})
	}
}

func TestClaudeClient_enhanceError_ModelNotFound(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"model not found", "model not found"},
		{"invalid model", "invalid model"},
		{"not_found_error", "not_found_error: model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), `model "claude-sonnet-4-6" not found`) {
				t.Fatalf("expected model not found error, got: %v", err)
			}
		})
	}
}

func TestClaudeClient_enhanceError_ContextLength(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"context length", "context length exceeded"},
		{"context window", "context window exceeded"},
		{"prompt too long", "prompt is too long"},
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

func TestClaudeClient_enhanceError_ServerError(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"500", "500 internal server error"},
		{"502", "502 bad gateway"},
		{"503", "503 service unavailable"},
		{"internal server error", "internal server error"},
		{"api_error", "api_error: something went wrong"},
		{"overloaded_error", "overloaded_error: too busy"},
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

func TestClaudeClient_enhanceError_NetworkError(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)

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

func TestClaudeClient_enhanceError_Unknown(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)
	err := client.enhanceError(errors.New("some random error"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "claude error:") {
		t.Fatalf("expected generic claude error, got: %v", err)
	}
}

func TestClaudeClient_enhanceError_CaseInsensitive(t *testing.T) {
	client := NewClaudeClient("sk-test", "claude-sonnet-4-6", 0)

	err := client.enhanceError(errors.New("TIMEOUT"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error for uppercase, got: %v", err)
	}
}

func TestClaudeClient_ProviderInterface(t *testing.T) {
	var _ Provider = (*ClaudeClient)(nil)
}
