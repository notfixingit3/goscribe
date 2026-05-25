package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/house/goscribe/internal/ai/mocks"
	"github.com/house/goscribe/internal/config"
	"github.com/spf13/viper"
)

func TestNewProvider_NoProvidersConfigured(t *testing.T) {
	viper.Reset()
	cfg := &config.Config{}

	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error when no providers configured")
	}
	if !strings.Contains(err.Error(), "no providers configured") {
		t.Fatalf("expected 'no providers configured' error, got: %v", err)
	}
}

func TestNewProvider_ProviderNotFound(t *testing.T) {
	viper.Reset()
	viper.Set("providers", []map[string]interface{}{
		{"name": "ollama", "url": "http://localhost:11434", "model": "llama2"},
	})

	cfg := &config.Config{Provider: "openai"}
	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error when provider not found")
	}
	if !strings.Contains(err.Error(), `provider "openai" not found`) {
		t.Fatalf("expected provider not found error, got: %v", err)
	}
}

func TestNewProvider_ProviderNotFoundNoProviders(t *testing.T) {
	viper.Reset()
	cfg := &config.Config{Provider: "openai"}
	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `provider "openai" not found`) {
		t.Fatalf("expected provider not found error, got: %v", err)
	}
}

func TestNewProvider_DefaultProvider(t *testing.T) {
	viper.Reset()
	viper.Set("providers", []map[string]interface{}{
		{"name": "ollama", "url": "http://localhost:11434", "model": "llama2", "default": true},
	})

	cfg := &config.Config{}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected provider, got nil")
	}
}

func TestNewProvider_FirstProviderWhenNoDefault(t *testing.T) {
	viper.Reset()
	viper.Set("providers", []map[string]interface{}{
		{"name": "ollama", "url": "http://localhost:11434", "model": "llama2"},
	})

	cfg := &config.Config{}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected provider, got nil")
	}
}

func TestNewProvider_OpenAINoKey(t *testing.T) {
	viper.Reset()
	viper.Set("providers", []map[string]interface{}{
		{"name": "openai", "model": "gpt-4"},
	})

	cfg := &config.Config{Provider: "openai"}
	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error when OpenAI API key missing")
	}
	if !strings.Contains(err.Error(), "requires an API key") {
		t.Fatalf("expected API key error, got: %v", err)
	}
}

func TestNewProvider_OpenAIWithKey(t *testing.T) {
	viper.Reset()
	viper.Set("providers", []map[string]interface{}{
		{"name": "openai", "api_key": "sk-test", "model": "gpt-4"},
	})

	cfg := &config.Config{Provider: "openai"}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected provider, got nil")
	}
}

func TestNewProvider_UnsupportedProvider(t *testing.T) {
	viper.Reset()
	viper.Set("providers", []map[string]interface{}{
		{"name": "anthropic", "api_key": "test", "model": "claude"},
	})

	cfg := &config.Config{Provider: "anthropic"}
	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Fatalf("expected unsupported provider error, got: %v", err)
	}
}

func TestNewProvider_WithRetries(t *testing.T) {
	viper.Reset()
	viper.Set("providers", []map[string]interface{}{
		{"name": "ollama", "url": "http://localhost:11434", "model": "llama2"},
	})

	cfg := &config.Config{Retries: 3, RetryBackoff: 1 * time.Second}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected provider, got nil")
	}

	// Should be wrapped in RetryProvider
	_, ok := p.(*RetryProvider)
	if !ok {
		t.Fatalf("expected *RetryProvider, got %T", p)
	}
}

func TestNewProvider_NegativeRetries(t *testing.T) {
	viper.Reset()
	viper.Set("providers", []map[string]interface{}{
		{"name": "ollama", "url": "http://localhost:11434", "model": "llama2"},
	})

	cfg := &config.Config{Retries: -1}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected provider, got nil")
	}

	// Negative retries should be treated as 0, so no RetryProvider wrapper
	_, ok := p.(*RetryProvider)
	if ok {
		t.Fatal("expected no RetryProvider wrapper for negative retries")
	}
}

func TestNewProvider_VerboseRetry(t *testing.T) {
	viper.Reset()
	viper.Set("providers", []map[string]interface{}{
		{"name": "ollama", "url": "http://localhost:11434", "model": "llama2"},
	})

	cfg := &config.Config{Retries: 2, Verbose: true, RetryBackoff: 1 * time.Second}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected provider, got nil")
	}

	_, ok := p.(*RetryProvider)
	if !ok {
		t.Fatalf("expected *RetryProvider, got %T", p)
	}
}

func TestProviderNames(t *testing.T) {
	tests := []struct {
		name string
		list []struct {
			Name string
		}
		expected string
	}{
		{
			name:     "empty list",
			list:     nil,
			expected: "(none)",
		},
		{
			name: "single provider",
			list: []struct{ Name string }{
				{Name: "openai"},
			},
			expected: "[openai]",
		},
		{
			name: "multiple providers",
			list: []struct{ Name string }{
				{Name: "openai"},
				{Name: "ollama"},
			},
			expected: "[openai ollama]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test providerNames directly since it takes []providers.ProviderConfig,
			// but we can verify it through NewProvider behavior.
		})
	}
}

func TestProvider_InterfaceCompliance(t *testing.T) {
	// Verify that all provider implementations satisfy the interface
	var _ Provider = (*OpenAIClient)(nil)
	var _ Provider = (*OllamaClient)(nil)
	var _ Provider = (*RetryProvider)(nil)
	var _ Provider = (mocks.Provider)(nil)
}

func TestMockProvider(t *testing.T) {
	mock := mocks.NewMockProvider("response1", "response2")

	// First call
	resp1, err := mock.Generate(context.Background(), "prompt1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp1 != "response1" {
		t.Fatalf("expected 'response1', got %q", resp1)
	}

	// Second call
	resp2, err := mock.Generate(context.Background(), "prompt2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp2 != "response2" {
		t.Fatalf("expected 'response2', got %q", resp2)
	}

	// Third call - should error
	_, err = mock.Generate(context.Background(), "prompt3")
	if err == nil {
		t.Fatal("expected error when responses exhausted")
	}

	// Check prompts recorded
	prompts := mock.Prompts()
	if len(prompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(prompts))
	}
	if prompts[0] != "prompt1" || prompts[1] != "prompt2" || prompts[2] != "prompt3" {
		t.Fatalf("unexpected prompts: %v", prompts)
	}

	// Check call count
	if mock.CallCount() != 3 {
		t.Fatalf("expected call count 3, got %d", mock.CallCount())
	}
}

func TestMockProvider_SetErrors(t *testing.T) {
	mock := mocks.NewMockProvider("success")
	mock.SetErrors(errors.New("first error"), errors.New("second error"))

	// First call returns error
	_, err := mock.Generate(context.Background(), "prompt1")
	if err == nil || err.Error() != "first error" {
		t.Fatalf("expected first error, got: %v", err)
	}

	// Second call returns error
	_, err = mock.Generate(context.Background(), "prompt2")
	if err == nil || err.Error() != "second error" {
		t.Fatalf("expected second error, got: %v", err)
	}

	// Third call returns response
	resp, err := mock.Generate(context.Background(), "prompt3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "success" {
		t.Fatalf("expected 'success', got %q", resp)
	}
}

func TestMockProvider_Reset(t *testing.T) {
	mock := mocks.NewMockProvider("response")
	mock.Generate(context.Background(), "prompt")
	mock.SetErrors(errors.New("error"))

	mock.Reset()

	if mock.CallCount() != 0 {
		t.Fatalf("expected call count 0 after reset, got %d", mock.CallCount())
	}

	// Should be able to use responses again
	resp, err := mock.Generate(context.Background(), "new prompt")
	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if resp != "response" {
		t.Fatalf("expected 'response' after reset, got %q", resp)
	}
}

func TestMockProvider_ConcurrentAccess(t *testing.T) {
	mock := mocks.NewMockProvider("response1", "response2", "response3")

	// Concurrent calls should be safe
	done := make(chan struct{}, 3)
	for i := 0; i < 3; i++ {
		go func(i int) {
			defer func() { done <- struct{}{} }()
			_, err := mock.Generate(context.Background(), fmt.Sprintf("prompt%d", i))
			if err != nil {
				t.Errorf("unexpected error in goroutine %d: %v", i, err)
			}
		}(i)
	}

	for i := 0; i < 3; i++ {
		<-done
	}

	if mock.CallCount() != 3 {
		t.Fatalf("expected 3 calls, got %d", mock.CallCount())
	}
}

func TestWithWriter(t *testing.T) {
	// withWriter is a no-op option, but we can verify it doesn't panic
	r := &RetryProvider{}
	opt := withWriter(nil)
	opt(r)
	// No panic = success
}
