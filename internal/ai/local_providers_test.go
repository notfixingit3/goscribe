package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestNewLMStudioClient_DefaultValues(t *testing.T) {
	client := NewLMStudioClient("", "")
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.url != "http://localhost:1234" {
		t.Fatalf("expected default URL 'http://localhost:1234', got %q", client.url)
	}
	if client.openAI.model != "local-model" {
		t.Fatalf("expected default model 'local-model', got %q", client.openAI.model)
	}
}

func TestNewLMStudioClient_CustomValues(t *testing.T) {
	client := NewLMStudioClient("http://192.168.1.100:1234", "custom-model")
	if client.url != "http://192.168.1.100:1234" {
		t.Fatalf("expected URL 'http://192.168.1.100:1234', got %q", client.url)
	}
	if client.openAI.model != "custom-model" {
		t.Fatalf("expected model 'custom-model', got %q", client.openAI.model)
	}
}

func TestLMStudioClient_enhanceError_ConnectionRefused(t *testing.T) {
	client := NewLMStudioClient("http://localhost:1234", "test-model")
	err := client.enhanceError(fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "LM Studio") {
		t.Fatalf("expected LM Studio error, got: %v", err)
	}
}

func TestLMStudioClient_enhanceError_ModelNotLoaded(t *testing.T) {
	client := NewLMStudioClient("http://localhost:1234", "test-model")
	err := client.enhanceError(fmt.Errorf("model not loaded"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "model") {
		t.Fatalf("expected model error, got: %v", err)
	}
}

func TestNewLLaMACppClient_DefaultValues(t *testing.T) {
	client := NewLLaMACppClient("", "")
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.url != "http://localhost:8080" {
		t.Fatalf("expected default URL 'http://localhost:8080', got %q", client.url)
	}
	if client.openAI.model != "llama2" {
		t.Fatalf("expected default model 'llama2', got %q", client.openAI.model)
	}
}

func TestNewLLaMACppClient_CustomValues(t *testing.T) {
	client := NewLLaMACppClient("http://192.168.1.100:8080", "mistral")
	if client.url != "http://192.168.1.100:8080" {
		t.Fatalf("expected URL 'http://192.168.1.100:8080', got %q", client.url)
	}
	if client.openAI.model != "mistral" {
		t.Fatalf("expected model 'mistral', got %q", client.openAI.model)
	}
}

func TestLLaMACppClient_enhanceError_ConnectionRefused(t *testing.T) {
	client := NewLLaMACppClient("http://localhost:8080", "test-model")
	err := client.enhanceError(fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "LLaMA.cpp") {
		t.Fatalf("expected LLaMA.cpp error, got: %v", err)
	}
}

func TestNewOllamaCloudClient_DefaultValues(t *testing.T) {
	client, err := NewOllamaCloudClient("", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.apiKey != "" {
		t.Fatalf("expected empty API key, got %q", client.apiKey)
	}
}

func TestNewOllamaCloudClient_WithAPIKey(t *testing.T) {
	client, err := NewOllamaCloudClient("sk-test", "https://cloud.ollama.com", "llama2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.apiKey != "sk-test" {
		t.Fatalf("expected API key 'sk-test', got %q", client.apiKey)
	}
}

func TestOllamaCloudClient_enhanceError_Auth(t *testing.T) {
	client, _ := NewOllamaCloudClient("", "", "")
	err := client.enhanceError(fmt.Errorf("unauthorized"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "authentication") {
		t.Fatalf("expected auth error, got: %v", err)
	}
}

func TestNewOllamaSelfHostedClient_RequiresURL(t *testing.T) {
	_, err := NewOllamaSelfHostedClient("", "llama2")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
	if !strings.Contains(err.Error(), "requires a URL") {
		t.Fatalf("expected URL required error, got: %v", err)
	}
}

func TestNewOllamaSelfHostedClient_WithURL(t *testing.T) {
	client, err := NewOllamaSelfHostedClient("http://192.168.1.100:11434", "llama2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client, got nil")
	}
}

func TestOllamaSelfHostedClient_enhanceError_Connection(t *testing.T) {
	client, _ := NewOllamaSelfHostedClient("http://192.168.1.100:11434", "llama2")
	err := client.enhanceError(context.DeadlineExceeded)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "self-hosted") {
		t.Fatalf("expected self-hosted error, got: %v", err)
	}
}

func TestNewOpenCodeClient_DefaultValues(t *testing.T) {
	client := NewOpenCodeClient("", "")
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.url != "http://localhost:8080" {
		t.Fatalf("expected default URL 'http://localhost:8080', got %q", client.url)
	}
	if client.openAI.model != "default" {
		t.Fatalf("expected default model 'default', got %q", client.openAI.model)
	}
}

func TestNewOpenCodeClient_CustomValues(t *testing.T) {
	client := NewOpenCodeClient("http://agent.local:9090", "gpt-4")
	if client.url != "http://agent.local:9090" {
		t.Fatalf("expected URL 'http://agent.local:9090', got %q", client.url)
	}
	if client.openAI.model != "gpt-4" {
		t.Fatalf("expected model 'gpt-4', got %q", client.openAI.model)
	}
}

func TestOpenCodeClient_enhanceError_ConnectionRefused(t *testing.T) {
	client := NewOpenCodeClient("http://localhost:8080", "test-model")
	err := client.enhanceError(fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "OpenCode") {
		t.Fatalf("expected OpenCode error, got: %v", err)
	}
}
