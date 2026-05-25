package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewOllamaClient_DefaultModel(t *testing.T) {
	client, err := NewOllamaClient("http://localhost:11434", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.model != "llama2" {
		t.Fatalf("expected default model 'llama2', got %q", client.model)
	}
}

func TestNewOllamaClient_CustomModel(t *testing.T) {
	client, err := NewOllamaClient("http://localhost:11434", "mistral")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.model != "mistral" {
		t.Fatalf("expected model 'mistral', got %q", client.model)
	}
}

func TestNewOllamaClient_InvalidURL(t *testing.T) {
	_, err := NewOllamaClient("://invalid-url", "llama2")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "invalid Ollama URL") {
		t.Fatalf("expected invalid URL error, got: %v", err)
	}
}

func TestNewOllamaClient_MissingScheme(t *testing.T) {
	_, err := NewOllamaClient("://missing-scheme", "llama2")
	if err == nil {
		t.Fatal("expected error for missing scheme")
	}
	if !strings.Contains(err.Error(), "invalid Ollama URL") {
		t.Fatalf("expected invalid URL error, got: %v", err)
	}
}

func TestNewOllamaClient_EmptyURL(t *testing.T) {
	// When URL is empty, it tries to create client from environment
	// This may or may not succeed depending on environment
	client, err := NewOllamaClient("", "llama2")
	// We just verify it doesn't panic and either succeeds or returns an error
	if err != nil {
		// Error is acceptable - Ollama might not be configured in environment
		if !strings.Contains(err.Error(), "could not create Ollama client from environment") {
			t.Fatalf("unexpected error: %v", err)
		}
	} else if client == nil {
		t.Fatal("expected client or error, got nil client without error")
	}
}

func TestOllamaClient_Generate_ContextCancellation(t *testing.T) {
	client, err := NewOllamaClient("http://localhost:11434", "llama2")
	if err != nil {
		// Skip if Ollama client can't be created (e.g., no environment)
		t.Skipf("skipping: could not create Ollama client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = client.Generate(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error due to canceled context")
	}
}

func TestOllamaClient_enhanceError_Nil(t *testing.T) {
	client := &OllamaClient{model: "llama2"}
	err := client.enhanceError(nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestOllamaClient_enhanceError_ConnectionRefused(t *testing.T) {
	client := &OllamaClient{model: "llama2"}
	err := client.enhanceError(errors.New("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cannot connect to Ollama") {
		t.Fatalf("expected connection refused error, got: %v", err)
	}
}

func TestOllamaClient_enhanceError_ModelNotFound(t *testing.T) {
	client := &OllamaClient{model: "llama2"}

	tests := []struct {
		name   string
		errMsg string
	}{
		{"model not found", "model not found"},
		{"no such model", "no such model"},
		{"404", "404 not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), `model "llama2" not found`) {
				t.Fatalf("expected model not found error, got: %v", err)
			}
		})
	}
}

func TestOllamaClient_enhanceError_Timeout(t *testing.T) {
	client := &OllamaClient{model: "llama2"}

	tests := []struct {
		name   string
		errMsg string
	}{
		{"deadline exceeded", "context deadline exceeded"},
		{"timeout", "request timeout"},
		{"i/o timeout", "i/o timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "timed out") {
				t.Fatalf("expected timeout error, got: %v", err)
			}
		})
	}
}

func TestOllamaClient_enhanceError_ContextCancelled(t *testing.T) {
	client := &OllamaClient{model: "llama2"}

	tests := []struct {
		name   string
		errMsg string
	}{
		{"context canceled", "context canceled"},
		{"context cancelled", "context cancelled"}, //nolint:misspell // testing Ollama's British spelling
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "canceled") {
				t.Fatalf("expected cancellation error, got: %v", err)
			}
		})
	}
}

func TestOllamaClient_enhanceError_ServerError(t *testing.T) {
	client := &OllamaClient{model: "llama2"}
	err := client.enhanceError(errors.New("500 internal server error"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Fatalf("expected server error, got: %v", err)
	}
}

func TestOllamaClient_enhanceError_NetworkError(t *testing.T) {
	client := &OllamaClient{model: "llama2"}

	tests := []struct {
		name   string
		errMsg string
	}{
		{"network", "network is unreachable"},
		{"eof", "unexpected EOF"},
		{"connection reset", "connection reset by peer"},
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

func TestOllamaClient_enhanceError_Unknown(t *testing.T) {
	client := &OllamaClient{model: "llama2"}
	err := client.enhanceError(errors.New("some random error"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "ollama error:") {
		t.Fatalf("expected generic ollama error, got: %v", err)
	}
}

func TestOllamaClient_enhanceError_CaseInsensitive(t *testing.T) {
	client := &OllamaClient{model: "llama2"}

	// Test that matching is case-insensitive
	err := client.enhanceError(errors.New("CONNECTION REFUSED"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cannot connect to Ollama") {
		t.Fatalf("expected connection error for uppercase, got: %v", err)
	}
}

func TestOllamaClient_ProviderInterface(t *testing.T) {
	var _ Provider = (*OllamaClient)(nil)
}
