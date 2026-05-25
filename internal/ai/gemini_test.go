package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewGeminiClient_DefaultModel(t *testing.T) {
	client := NewGeminiClient("test-key", "", 0)
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.model != "gemini-2.0-flash" {
		t.Fatalf("expected default model 'gemini-2.0-flash', got %q", client.model)
	}
}

func TestNewGeminiClient_CustomModel(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-pro", 0)
	if client.model != "gemini-pro" {
		t.Fatalf("expected model 'gemini-pro', got %q", client.model)
	}
}

func TestNewGeminiClient_WithTimeout(t *testing.T) {
	timeout := 30 * time.Second
	client := NewGeminiClient("test-key", "gemini-2.0-flash", timeout)
	if client.timeout != timeout {
		t.Fatalf("expected timeout %v, got %v", timeout, client.timeout)
	}
}

func TestNewGeminiClient_DefaultTimeout(t *testing.T) {
	client := NewGeminiClient("test-key", "", 0)
	if client.timeout != 5*time.Minute {
		t.Fatalf("expected default timeout 5m, got %v", client.timeout)
	}
}

func TestNewGeminiClient_EnvFallback(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "env-key-123")
	client := NewGeminiClient("", "gemini-2.0-flash", 0)
	if client.apiKey != "env-key-123" {
		t.Fatalf("expected API key from env, got %q", client.apiKey)
	}
}

func TestNewGeminiClient_ExplicitKeyOverridesEnv(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "env-key")
	client := NewGeminiClient("explicit-key", "gemini-2.0-flash", 0)
	if client.apiKey != "explicit-key" {
		t.Fatalf("expected explicit API key, got %q", client.apiKey)
	}
}

func TestGeminiClient_Generate_MissingAPIKey(t *testing.T) {
	client := NewGeminiClient("", "gemini-2.0-flash", 0)
	client.apiKey = ""

	_, err := client.Generate(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "API key is required") {
		t.Fatalf("expected API key error, got: %v", err)
	}
}

func TestGeminiClient_Generate_Success(t *testing.T) {
	wantText := "Generated documentation content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "models/gemini-2.0-flash:generateContent") {
			t.Errorf("unexpected URL path: %s", r.URL.Path)
		}

		resp := geminiResponse{
			Candidates: []geminiCandidate{
				{
					Content: geminiContent{
						Parts: []geminiPart{{Text: wantText}},
					},
					FinishReason: "STOP",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.0-flash", 10*time.Second)
	client.baseURL = server.URL

	result, err := client.Generate(context.Background(), "generate docs for main.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != wantText {
		t.Fatalf("expected %q, got %q", wantText, result)
	}
}

func TestGeminiClient_Generate_SafetyBlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiResponse{
			Candidates: []geminiCandidate{
				{
					Content:      geminiContent{},
					FinishReason: "SAFETY",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.0-flash", 10*time.Second)
	client.baseURL = server.URL

	_, err := client.Generate(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected error for safety block")
	}
	if !strings.Contains(err.Error(), "safety filters") {
		t.Fatalf("expected safety block error, got: %v", err)
	}
}

func TestGeminiClient_Generate_NoCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiResponse{Candidates: nil}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.0-flash", 10*time.Second)
	client.baseURL = server.URL

	_, err := client.Generate(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected error for no candidates")
	}
	if !strings.Contains(err.Error(), "no candidates") {
		t.Fatalf("expected no candidates error, got: %v", err)
	}
}

func TestGeminiClient_Generate_EmptyParts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiResponse{
			Candidates: []geminiCandidate{
				{
					Content:      geminiContent{Parts: nil},
					FinishReason: "STOP",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.0-flash", 10*time.Second)
	client.baseURL = server.URL

	_, err := client.Generate(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	if !strings.Contains(err.Error(), "empty content") {
		t.Fatalf("expected empty content error, got: %v", err)
	}
}

func TestGeminiClient_Generate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiResponse{
			Error: &geminiError{
				Code:    400,
				Message: "API key not valid",
				Status:  "INVALID_ARGUMENT",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.0-flash", 10*time.Second)
	client.baseURL = server.URL

	_, err := client.Generate(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Fatalf("expected auth error, got: %v", err)
	}
}

func TestGeminiClient_Generate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		resp := geminiResponse{
			Error: &geminiError{
				Code:    404,
				Message: "model gemini-xyz not found",
				Status:  "NOT_FOUND",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-xyz", 10*time.Second)
	client.baseURL = server.URL

	_, err := client.Generate(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected model not found error, got: %v", err)
	}
}

func TestGeminiClient_Generate_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.0-flash", 1*time.Millisecond)
	client.baseURL = server.URL

	_, err := client.Generate(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestGeminiClient_Generate_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.0-flash", 10*time.Second)
	client.baseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Generate(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error due to canceled context")
	}
}

func TestGeminiClient_enhanceError_Nil(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)
	err := client.enhanceError(nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestGeminiClient_enhanceError_Timeout(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 30*time.Second)
	err := client.enhanceError(errors.New("request deadline exceeded"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestGeminiClient_enhanceError_Auth(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)

	tests := []struct {
		name    string
		errMsg  string
		wantSub string
	}{
		{"unauthorized", "unauthorized", "authentication failed"},
		{"invalid api key", "invalid api key provided", "authentication failed"},
		{"api key not valid", "API key not valid", "authentication failed"},
		{"permission denied", "permission_denied", "authentication failed"},
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

func TestGeminiClient_enhanceError_RateLimit(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"rate limit", "rate limit exceeded"},
		{"429", "429 too many requests"},
		{"resource exhausted", "resource_exhausted"},
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

func TestGeminiClient_enhanceError_ModelNotFound(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"model not found", "model not found"},
		{"not_found", "not_found"},
		{"does not exist", "model does not exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.enhanceError(errors.New(tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "not found") {
				t.Fatalf("expected model not found error, got: %v", err)
			}
		})
	}
}

func TestGeminiClient_enhanceError_ContextLength(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)
	err := client.enhanceError(errors.New("exceeds token limit"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "context length exceeded") {
		t.Fatalf("expected context length error, got: %v", err)
	}
}

func TestGeminiClient_enhanceError_Safety(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)
	err := client.enhanceError(errors.New("blocked by safety"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "safety filters") {
		t.Fatalf("expected safety error, got: %v", err)
	}
}

func TestGeminiClient_enhanceError_ServerError(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"500", "500 internal server error"},
		{"502", "502 bad gateway"},
		{"503", "503 service unavailable"},
		{"internal_error", "internal_error"},
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

func TestGeminiClient_enhanceError_NetworkError(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)

	tests := []struct {
		name   string
		errMsg string
	}{
		{"connection refused", "connection refused"},
		{"network", "network is unreachable"},
		{"eof", "unexpected EOF"},
		{"dns", "dns resolution failed"},
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

func TestGeminiClient_enhanceError_Unknown(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)
	err := client.enhanceError(errors.New("some random error"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "gemini error:") {
		t.Fatalf("expected generic gemini error, got: %v", err)
	}
}

func TestGeminiClient_enhanceError_CaseInsensitive(t *testing.T) {
	client := NewGeminiClient("test-key", "gemini-2.0-flash", 0)
	err := client.enhanceError(errors.New("TIMEOUT"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error for uppercase, got: %v", err)
	}
}

func TestGeminiClient_ProviderInterface(t *testing.T) {
	var _ Provider = (*GeminiClient)(nil)
}
