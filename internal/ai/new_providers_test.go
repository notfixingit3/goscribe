package ai

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewXAIClient_DefaultValues(t *testing.T) {
	client := NewXAIClient("sk-test", "", "")
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.url != "https://api.x.ai" {
		t.Fatalf("expected default URL 'https://api.x.ai', got %q", client.url)
	}
	if client.openAI.model != "grok-2" {
		t.Fatalf("expected default model 'grok-2', got %q", client.openAI.model)
	}
}

func TestNewXAIClient_CustomValues(t *testing.T) {
	client := NewXAIClient("sk-test", "https://custom.x.ai", "grok-2-mini")
	if client.url != "https://custom.x.ai" {
		t.Fatalf("expected URL 'https://custom.x.ai', got %q", client.url)
	}
	if client.openAI.model != "grok-2-mini" {
		t.Fatalf("expected model 'grok-2-mini', got %q", client.openAI.model)
	}
}

func TestXAIClient_enhanceError_ConnectionRefused(t *testing.T) {
	client := NewXAIClient("sk-test", "https://api.x.ai", "grok-2")
	err := client.enhanceError(fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "xAI") {
		t.Fatalf("expected xAI error, got: %v", err)
	}
}

func TestNewNVIDIAClient_DefaultValues(t *testing.T) {
	client := NewNVIDIAClient("nvapi-test", "", "")
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.url != "https://integrate.api.nvidia.com" {
		t.Fatalf("expected default URL 'https://integrate.api.nvidia.com', got %q", client.url)
	}
	if client.openAI.model != "meta/llama-3.1-405b-instruct" {
		t.Fatalf("expected default model 'meta/llama-3.1-405b-instruct', got %q", client.openAI.model)
	}
}

func TestNewNVIDIAClient_CustomValues(t *testing.T) {
	client := NewNVIDIAClient("nvapi-test", "https://custom.nvidia.com", "mixtral-8x22b")
	if client.url != "https://custom.nvidia.com" {
		t.Fatalf("expected URL 'https://custom.nvidia.com', got %q", client.url)
	}
	if client.openAI.model != "mixtral-8x22b" {
		t.Fatalf("expected model 'mixtral-8x22b', got %q", client.openAI.model)
	}
}

func TestNVIDIAClient_enhanceError_ConnectionRefused(t *testing.T) {
	client := NewNVIDIAClient("nvapi-test", "https://integrate.api.nvidia.com", "test-model")
	err := client.enhanceError(fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "NVIDIA") {
		t.Fatalf("expected NVIDIA error, got: %v", err)
	}
}

func TestNewGitHubCopilotClient_DefaultValues(t *testing.T) {
	client := NewGitHubCopilotClient("ghp-test", "", "")
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.url != "https://api.githubcopilot.com" {
		t.Fatalf("expected default URL 'https://api.githubcopilot.com', got %q", client.url)
	}
	if client.openAI.model != "gpt-4" {
		t.Fatalf("expected default model 'gpt-4', got %q", client.openAI.model)
	}
}

func TestNewGitHubCopilotClient_CustomValues(t *testing.T) {
	client := NewGitHubCopilotClient("ghp-test", "https://custom.copilot.com", "gpt-4o")
	if client.url != "https://custom.copilot.com" {
		t.Fatalf("expected URL 'https://custom.copilot.com', got %q", client.url)
	}
	if client.openAI.model != "gpt-4o" {
		t.Fatalf("expected model 'gpt-4o', got %q", client.openAI.model)
	}
}

func TestGitHubCopilotClient_enhanceError_ConnectionRefused(t *testing.T) {
	client := NewGitHubCopilotClient("ghp-test", "https://api.githubcopilot.com", "test-model")
	err := client.enhanceError(fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "GitHub Copilot") {
		t.Fatalf("expected GitHub Copilot error, got: %v", err)
	}
}

func TestNewZaiClient_DefaultValues(t *testing.T) {
	client := NewZaiClient("zai-test", "", "")
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.url != "https://api.z.ai" {
		t.Fatalf("expected default URL 'https://api.z.ai', got %q", client.url)
	}
	if client.openAI.model != "z-large" {
		t.Fatalf("expected default model 'z-large', got %q", client.openAI.model)
	}
}

func TestNewZaiClient_CustomValues(t *testing.T) {
	client := NewZaiClient("zai-test", "https://custom.z.ai", "z-small")
	if client.url != "https://custom.z.ai" {
		t.Fatalf("expected URL 'https://custom.z.ai', got %q", client.url)
	}
	if client.openAI.model != "z-small" {
		t.Fatalf("expected model 'z-small', got %q", client.openAI.model)
	}
}

func TestZaiClient_enhanceError_ConnectionRefused(t *testing.T) {
	client := NewZaiClient("zai-test", "https://api.z.ai", "test-model")
	err := client.enhanceError(fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Z.ai") {
		t.Fatalf("expected Z.ai error, got: %v", err)
	}
}

func TestNewVLLMClient_DefaultValues(t *testing.T) {
	client := NewVLLMClient("", "")
	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.url != "http://localhost:8000" {
		t.Fatalf("expected default URL 'http://localhost:8000', got %q", client.url)
	}
	if client.openAI.model != "default" {
		t.Fatalf("expected default model 'default', got %q", client.openAI.model)
	}
}

func TestNewVLLMClient_CustomValues(t *testing.T) {
	client := NewVLLMClient("http://192.168.1.100:8000", "llama2")
	if client.url != "http://192.168.1.100:8000" {
		t.Fatalf("expected URL 'http://192.168.1.100:8000', got %q", client.url)
	}
	if client.openAI.model != "llama2" {
		t.Fatalf("expected model 'llama2', got %q", client.openAI.model)
	}
}

func TestVLLMClient_enhanceError_ConnectionRefused(t *testing.T) {
	client := NewVLLMClient("http://localhost:8000", "test-model")
	err := client.enhanceError(fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "vLLM") {
		t.Fatalf("expected vLLM error, got: %v", err)
	}
}

func TestVLLMClient_enhanceError_GPUError(t *testing.T) {
	client := NewVLLMClient("http://localhost:8000", "test-model")
	err := client.enhanceError(fmt.Errorf("CUDA out of memory"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "GPU") {
		t.Fatalf("expected GPU error, got: %v", err)
	}
}
