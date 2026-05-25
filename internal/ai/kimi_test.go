package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewKimiClient_DefaultModel(t *testing.T) {
	client := NewKimiClient("test-key", "", 0)
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.model != DefaultKimiModel {
		t.Fatalf("expected default model %q, got %q", DefaultKimiModel, client.model)
	}
	if client.client == nil {
		t.Fatal("expected Kimi client to be initialized")
	}
}

func TestNewKimiClient_CustomModel(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-32k", 0)
	if client.model != "moonshot-v1-32k" {
		t.Fatalf("expected model 'moonshot-v1-32k', got %q", client.model)
	}
}

func TestNewKimiClient_WithTimeout(t *testing.T) {
	timeout := 30 * time.Second
	client := NewKimiClient("test-key", "moonshot-v1-8k", timeout)
	if client.timeout != timeout {
		t.Fatalf("expected timeout %v, got %v", timeout, client.timeout)
	}
}

func TestKimiClient_Generate_ContextTimeout(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 1*time.Millisecond)

	ctx := context.Background()
	_, err := client.Generate(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error due to timeout")
	}
	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("expected timeout-related error, got: %v", err)
	}
}

func TestKimiClient_Generate_ContextCancellation(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Generate(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error due to canceled context")
	}
}

func TestKimiClient_enhanceError_Nil(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)
	err := client.enhanceError(nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestKimiClient_enhanceError_Timeout(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 30*time.Second)
	err := client.enhanceError(errors.New("request deadline exceeded"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestKimiClient_enhanceError_Auth(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)

	tests := []struct {
		name    string
		errMsg  string
		wantSub string
	}{
		{"unauthorized", "unauthorized", "authentication failed"},
		{"invalid api key", "invalid api key provided", "authentication failed"},
		{"incorrect api key", "incorrect api key", "authentication failed"},
		{"authentication error", "authentication error", "authentication failed"},
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

func TestKimiClient_enhanceError_RateLimit(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)

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

func TestKimiClient_enhanceError_Quota(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)
	err := client.enhanceError(errors.New("insufficient quota"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "quota exceeded") {
		t.Fatalf("expected quota error, got: %v", err)
	}
}

func TestKimiClient_enhanceError_ModelNotFound(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"model not found", "model not found"},
		{"invalid model", "invalid model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), `model "moonshot-v1-8k" not found`) {
				t.Fatalf("expected model not found error, got: %v", err)
			}
		})
	}
}

func TestKimiClient_enhanceError_ContextLength(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)

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

func TestKimiClient_enhanceError_ServerError(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)

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

func TestKimiClient_enhanceError_NetworkError(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)

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

func TestKimiClient_enhanceError_Unknown(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)
	err := client.enhanceError(errors.New("some random error"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "kimi error:") {
		t.Fatalf("expected generic kimi error, got: %v", err)
	}
}

func TestKimiClient_enhanceError_CaseInsensitive(t *testing.T) {
	client := NewKimiClient("test-key", "moonshot-v1-8k", 0)

	err := client.enhanceError(errors.New("TIMEOUT"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error for uppercase, got: %v", err)
	}
}

func TestKimiClient_ProviderInterface(t *testing.T) {
	var _ Provider = (*KimiClient)(nil)
}
