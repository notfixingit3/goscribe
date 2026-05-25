package goscribe

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate_WithMockProvider(t *testing.T) {
	mp := newMockProvider("# Documentation for main.go\n\nThis is a test doc.")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0600); err != nil {
		t.Fatalf("create source file: %v", err)
	}

	ctx := context.Background()
	result, err := client.Generate(ctx, srcDir, GenerateOptions{SkipStateSave: true})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result.SourcePath == "" {
		t.Fatal("expected SourcePath to be set")
	}
	if result.OutputDir != "docs" {
		t.Fatalf("expected OutputDir 'docs', got %q", result.OutputDir)
	}
	if result.StateSaved {
		t.Fatal("expected StateSaved to be false when SkipStateSave is true")
	}

	if mp.CallCount() != 1 {
		t.Fatalf("expected provider called once, got %d", mp.CallCount())
	}

	docFile := filepath.Join(result.OutputDir, "main.go.md")
	if _, err := os.Stat(docFile); os.IsNotExist(err) {
		t.Fatalf("expected doc file %s to exist", docFile)
	}
	content, err := os.ReadFile(docFile)
	if err != nil {
		t.Fatalf("read doc file: %v", err)
	}
	if string(content) != "# Documentation for main.go\n\nThis is a test doc." {
		t.Fatalf("unexpected doc content: %s", string(content))
	}
}

func TestGenerate_MultipleFiles(t *testing.T) {
	mp := newMockProvider("doc1", "doc2")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "a.go"), []byte("package a\n"), 0600); err != nil {
		t.Fatalf("create source file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "b.go"), []byte("package b\n"), 0600); err != nil {
		t.Fatalf("create source file: %v", err)
	}

	ctx := context.Background()
	result, err := client.Generate(ctx, srcDir, GenerateOptions{SkipStateSave: true})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if mp.CallCount() != 2 {
		t.Fatalf("expected provider called twice, got %d", mp.CallCount())
	}

	docA := filepath.Join(result.OutputDir, "a.go.md")
	docB := filepath.Join(result.OutputDir, "b.go.md")
	if _, err := os.Stat(docA); os.IsNotExist(err) {
		t.Fatalf("expected doc file %s to exist", docA)
	}
	if _, err := os.Stat(docB); os.IsNotExist(err) {
		t.Fatalf("expected doc file %s to exist", docB)
	}
}

func TestGenerate_PathValidation_NonExistent(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	_, err = client.Generate(ctx, "/nonexistent/path", GenerateOptions{})
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

func TestGenerate_PathValidation_FileNotDir(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "file.go")
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

func TestGenerate_OutputDirCreation(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(
		WithProvider(mp),
		WithOutputDir("custom-output"),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n"), 0600); err != nil {
		t.Fatalf("create source file: %v", err)
	}

	ctx := context.Background()
	result, err := client.Generate(ctx, srcDir, GenerateOptions{SkipStateSave: true})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result.OutputDir != "custom-output" {
		t.Fatalf("expected OutputDir 'custom-output', got %q", result.OutputDir)
	}

	info, err := os.Stat(result.OutputDir)
	if err != nil {
		t.Fatalf("stat output dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected output dir to be a directory")
	}
}

func TestGenerate_OutputDirOverride(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(
		WithProvider(mp),
		WithOutputDir("default-output"),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n"), 0600); err != nil {
		t.Fatalf("create source file: %v", err)
	}

	ctx := context.Background()
	result, err := client.Generate(ctx, srcDir, GenerateOptions{
		OutputDir:     "override-output",
		SkipStateSave: true,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result.OutputDir != "override-output" {
		t.Fatalf("expected OutputDir 'override-output', got %q", result.OutputDir)
	}
}

func TestGenerate_SkipsHiddenDirs(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "visible.go"), []byte("package visible\n"), 0600); err != nil {
		t.Fatalf("create source file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, ".hidden"), 0750); err != nil {
		t.Fatalf("create hidden dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, ".hidden", "hidden.go"), []byte("package hidden\n"), 0600); err != nil {
		t.Fatalf("create hidden file: %v", err)
	}

	ctx := context.Background()
	_, err = client.Generate(ctx, srcDir, GenerateOptions{SkipStateSave: true})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if mp.CallCount() != 1 {
		t.Fatalf("expected provider called once (only visible.go), got %d", mp.CallCount())
	}

	hiddenDoc := filepath.Join("docs", ".hidden", "hidden.go.md")
	if _, err := os.Stat(hiddenDoc); !os.IsNotExist(err) {
		t.Fatal("expected hidden file doc to NOT exist")
	}
}

func TestGenerate_ProviderError(t *testing.T) {
	mp := newMockProvider("doc")
	mp.SetErrors(errors.New("provider failure"))
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n"), 0600); err != nil {
		t.Fatalf("create source file: %v", err)
	}

	ctx := context.Background()
	_, err = client.Generate(ctx, srcDir, GenerateOptions{SkipStateSave: true})
	if err == nil {
		t.Fatal("expected error when provider fails")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrGenerationFailed) {
		t.Fatalf("expected ErrGenerationFailed, got %v", gerr.Kind)
	}
}

func TestGenerate_StateSaved(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n"), 0600); err != nil {
		t.Fatalf("create source file: %v", err)
	}

	ctx := context.Background()
	result, err := client.Generate(ctx, srcDir, GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result.StateSaved {
		t.Fatal("expected StateSaved to be false when not a git repo")
	}
}
