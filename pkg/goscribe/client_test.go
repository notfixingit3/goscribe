package goscribe

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/house/goscribe/internal/ai/mocks"
)

type mockProvider struct {
	*mocks.MockProvider
}

func newMockProvider(responses ...string) *mockProvider {
	return &mockProvider{MockProvider: mocks.NewMockProvider(responses...)}
}

func TestNewClient_WithProvider(t *testing.T) {
	mp := newMockProvider("doc1", "doc2")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient with provider failed: %v", err)
	}
	if client.provider == nil {
		t.Fatal("expected provider to be set")
	}
	if client.outputDir != "docs" {
		t.Fatalf("expected default outputDir 'docs', got %q", client.outputDir)
	}
	if client.timeout != 5*time.Minute {
		t.Fatalf("expected default timeout 5m, got %v", client.timeout)
	}
	if client.retries != 3 {
		t.Fatalf("expected default retries 3, got %d", client.retries)
	}
	if client.retryBackoff != 2*time.Second {
		t.Fatalf("expected default retryBackoff 2s, got %v", client.retryBackoff)
	}
}

func TestNewClient_WithAllOptions(t *testing.T) {
	mp := newMockProvider("response")
	client, err := NewClient(
		WithProvider(mp),
		WithOutputDir("custom-docs"),
		WithTimeout(30*time.Second),
		WithRetry(5, 1*time.Second),
		WithVerbose(true),
		WithModel("gpt-4"),
		WithConfiguredProvider("openai"),
	)
	if err != nil {
		t.Fatalf("NewClient with all options failed: %v", err)
	}
	if client.outputDir != "custom-docs" {
		t.Fatalf("expected outputDir 'custom-docs', got %q", client.outputDir)
	}
	if client.timeout != 30*time.Second {
		t.Fatalf("expected timeout 30s, got %v", client.timeout)
	}
	if client.retries != 5 {
		t.Fatalf("expected retries 5, got %d", client.retries)
	}
	if client.retryBackoff != 1*time.Second {
		t.Fatalf("expected retryBackoff 1s, got %v", client.retryBackoff)
	}
	if !client.verbose {
		t.Fatal("expected verbose to be true")
	}
	if client.model != "gpt-4" {
		t.Fatalf("expected model 'gpt-4', got %q", client.model)
	}
	if client.providerName != "openai" {
		t.Fatalf("expected providerName 'openai', got %q", client.providerName)
	}
}

func TestNewClient_NoProvider_WithoutConfig(t *testing.T) {
	tmpDir := t.TempDir()
	badConfig := filepath.Join(tmpDir, "nonexistent.yaml")
	_, err := NewClient(WithConfigFile(badConfig))
	if err == nil {
		t.Fatal("expected error when no provider and no valid config")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrProviderNotConfigured) && !errors.Is(gerr, ErrGenerationFailed) {
		t.Fatalf("expected ErrProviderNotConfigured or ErrGenerationFailed, got %v", gerr.Kind)
	}
}

func TestNewClient_NoProvider_WithUnsupportedProvider(t *testing.T) {
	_, err := NewClient(WithConfiguredProvider("nonexistent-provider"))
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrUnsupportedProvider) {
		t.Fatalf("expected ErrUnsupportedProvider, got %v", gerr.Kind)
	}
}

func TestClient_Generate_NonExistentPath(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	_, err = client.Generate(ctx, "/path/that/does/not/exist", GenerateOptions{})
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got %v", gerr.Kind)
	}
}

func TestClient_Generate_FileInsteadOfDirectory(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "notadir.go")
	if err := os.WriteFile(tmpFile, []byte("package main\n"), 0600); err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	ctx := context.Background()
	_, err = client.Generate(ctx, tmpFile, GenerateOptions{})
	if err == nil {
		t.Fatal("expected error when path is a file")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got %v", gerr.Kind)
	}
}

func TestClient_Update_MissingStateFile(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	tmpDir := t.TempDir()

	ctx := context.Background()
	_, err = client.Update(ctx, tmpDir, UpdateOptions{})
	if err == nil {
		t.Fatal("expected error when source is not a git repo")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrNotGitRepository) {
		t.Fatalf("expected ErrNotGitRepository, got %v", gerr.Kind)
	}
}

func TestClient_Update_NotGitRepository(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	tmpDir := t.TempDir()

	ctx := context.Background()
	_, err = client.Update(ctx, tmpDir, UpdateOptions{})
	if err == nil {
		t.Fatal("expected error when source is not a git repo")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrNotGitRepository) {
		t.Fatalf("expected ErrNotGitRepository, got %v", gerr.Kind)
	}
}

func TestClient_ContextCancellation_Generate(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.Generate(ctx, t.TempDir(), GenerateOptions{})
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrInvalidContext) {
		t.Fatalf("expected ErrInvalidContext, got %v", gerr.Kind)
	}
}

func TestClient_ContextCancellation_Update(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.Update(ctx, t.TempDir(), UpdateOptions{})
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrInvalidContext) {
		t.Fatalf("expected ErrInvalidContext, got %v", gerr.Kind)
	}
}

func TestClient_ContextTimeout_Generate(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	_, err = client.Generate(ctx, t.TempDir(), GenerateOptions{})
	if err == nil {
		t.Fatal("expected error for timed-out context")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrInvalidContext) {
		t.Fatalf("expected ErrInvalidContext, got %v", gerr.Kind)
	}
}

func TestClient_ConcurrentUsage(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	const numGoroutines = 20
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			_, err := client.Generate(ctx, t.TempDir(), GenerateOptions{SkipStateSave: true})
			if err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent Generate failed: %v", err)
	}
}

func TestClient_ConcurrentMixedUsage(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines*2)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			_, err := client.Generate(ctx, t.TempDir(), GenerateOptions{SkipStateSave: true})
			if err != nil {
				errCh <- err
			}
		}()
		go func() {
			defer wg.Done()
			ctx := context.Background()
			_, err := client.Update(ctx, t.TempDir(), UpdateOptions{})
			if err != nil {
				var gerr *Error
				if errors.As(err, &gerr) && errors.Is(gerr, ErrNotGitRepository) {
					return
				}
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent mixed usage failed unexpectedly: %v", err)
	}
}

func TestError_Is(t *testing.T) {
	err := &Error{Op: "test", Kind: ErrInvalidPath, Err: os.ErrNotExist}
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatal("expected Error.Is to match Kind")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal("expected Error.Is to match wrapped error")
	}
	if errors.Is(err, ErrGenerationFailed) {
		t.Fatal("expected Error.Is to NOT match unrelated error")
	}
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := &Error{Op: "test", Kind: ErrGenerationFailed, Err: inner}
	if !errors.Is(err, inner) {
		t.Fatal("expected errors.Is to find unwrapped error")
	}
}
