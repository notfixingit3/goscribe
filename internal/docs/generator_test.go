package docs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/house/goscribe/internal/ai/mocks"
)

func TestNewGenerator(t *testing.T) {
	provider := mocks.NewMockProvider("doc1")
	g := NewGenerator(provider, "output")
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.outputDir != "output" {
		t.Errorf("outputDir = %q, want %q", g.outputDir, "output")
	}
}

func TestGenerateCreatesOutputDir(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	// Create a source file
	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# Documentation")
	g := NewGenerator(provider, outputDir)

	if err := g.Generate(sourceDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory was not created")
	}
}

func TestGenerateProcessesGoFiles(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# Main Documentation")
	g := NewGenerator(provider, outputDir)

	if err := g.Generate(sourceDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Check output file was created
	outputFile := filepath.Join(outputDir, "main.go.md")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(content) != "# Main Documentation" {
		t.Errorf("output = %q, want %q", string(content), "# Main Documentation")
	}
}

func TestGenerateProcessesMarkdownFiles(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# README\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# README Doc")
	g := NewGenerator(provider, outputDir)

	if err := g.Generate(sourceDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	outputFile := filepath.Join(outputDir, "README.md.md")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("output file for markdown was not created")
	}
}

func TestGenerateSkipsNonSourceFiles(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "config.yaml"), []byte("key: value\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider()
	g := NewGenerator(provider, outputDir)

	if err := g.Generate(sourceDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if provider.CallCount() != 0 {
		t.Errorf("provider called %d times, want 0", provider.CallCount())
	}
}

func TestGenerateSkipsHiddenDirs(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	// Create hidden directory with a Go file
	hiddenDir := filepath.Join(sourceDir, ".git")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider()
	g := NewGenerator(provider, outputDir)

	if err := g.Generate(sourceDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if provider.CallCount() != 0 {
		t.Errorf("provider called %d times, want 0", provider.CallCount())
	}
}

func TestGenerateCreatesNestedOutputDirs(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	nestedDir := filepath.Join(sourceDir, "pkg", "utils")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "helper.go"), []byte("package utils\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# Helper Doc")
	g := NewGenerator(provider, outputDir)

	if err := g.Generate(sourceDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	outputFile := filepath.Join(outputDir, "pkg", "utils", "helper.go.md")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("nested output file was not created at %s", outputFile)
	}
}

func TestGenerateProviderError(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider()
	provider.SetErrors(errors.New("ai error"))
	g := NewGenerator(provider, outputDir)

	err := g.Generate(sourceDir)
	if err == nil {
		t.Fatal("expected error from provider")
	}
	if !strings.Contains(err.Error(), "ai error") {
		t.Errorf("error = %q, want to contain 'ai error'", err.Error())
	}
}

func TestGenerateMissingSourcePath(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "docs")
	provider := mocks.NewMockProvider("doc")
	g := NewGenerator(provider, outputDir)

	err := g.Generate("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing source path")
	}
}

func TestGenerateMultipleFiles(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "a.go"), []byte("package a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "b.go"), []byte("package b\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# A", "# B")
	g := NewGenerator(provider, outputDir)

	if err := g.Generate(sourceDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if provider.CallCount() != 2 {
		t.Errorf("provider called %d times, want 2", provider.CallCount())
	}

	// Verify prompts contain file names
	prompts := provider.Prompts()
	for _, prompt := range prompts {
		if !strings.Contains(prompt, "File: ") {
			t.Error("prompt missing file name")
		}
	}
}

func TestGeneratePreservesFileContentInPrompt(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	content := "package main\n\nfunc Hello() string {\n\treturn \"hello\"\n}\n"
	if err := os.WriteFile(filepath.Join(sourceDir, "hello.go"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# Hello Doc")
	g := NewGenerator(provider, outputDir)

	if err := g.Generate(sourceDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	prompts := provider.Prompts()
	if len(prompts) != 1 {
		t.Fatalf("got %d prompts, want 1", len(prompts))
	}
	if !strings.Contains(prompts[0], content) {
		t.Error("prompt does not contain file content")
	}
}

func TestCollectSourceFilesEmptyDir(t *testing.T) {
	sourceDir := t.TempDir()
	provider := mocks.NewMockProvider()
	g := NewGenerator(provider, "output")

	files, err := g.collectSourceFiles(sourceDir)
	if err != nil {
		t.Fatalf("collectSourceFiles: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("got %d files, want 0", len(files))
	}
}

func TestCollectSourceFilesReturnsRelativePaths(t *testing.T) {
	sourceDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider()
	g := NewGenerator(provider, "output")

	files, err := g.collectSourceFiles(sourceDir)
	if err != nil {
		t.Fatalf("collectSourceFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}
	if files[0] != "main.go" {
		t.Errorf("file = %q, want %q", files[0], "main.go")
	}
}
