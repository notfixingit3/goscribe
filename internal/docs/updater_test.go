package docs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/house/goscribe/internal/ai/mocks"
)

func TestNewUpdater(t *testing.T) {
	provider := mocks.NewMockProvider("doc")
	u := NewUpdater(provider, "output")
	if u == nil {
		t.Fatal("NewUpdater returned nil")
	}
	if u.outputDir != "output" {
		t.Errorf("outputDir = %q, want %q", u.outputDir, "output")
	}
}

func TestUpdateProcessesChangedFiles(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# Updated Main Doc")
	u := NewUpdater(provider, outputDir)

	if err := u.Update(sourceDir, []string{"main.go"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	outputFile := filepath.Join(outputDir, "main.go.md")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(content) != "# Updated Main Doc" {
		t.Errorf("output = %q, want %q", string(content), "# Updated Main Doc")
	}
}

func TestUpdateMultipleFiles(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "a.go"), []byte("package a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "b.go"), []byte("package b\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# Updated A", "# Updated B")
	u := NewUpdater(provider, outputDir)

	if err := u.Update(sourceDir, []string{"a.go", "b.go"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if provider.CallCount() != 2 {
		t.Errorf("provider called %d times, want 2", provider.CallCount())
	}

	// Verify both output files exist
	for _, file := range []string{"a.go.md", "b.go.md"} {
		path := filepath.Join(outputDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("output file %s was not created", file)
		}
	}
}

func TestUpdateEmptyChangedFiles(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	provider := mocks.NewMockProvider()
	u := NewUpdater(provider, outputDir)

	if err := u.Update(sourceDir, []string{}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if provider.CallCount() != 0 {
		t.Errorf("provider called %d times, want 0", provider.CallCount())
	}
}

func TestUpdateMissingSourceFile(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	provider := mocks.NewMockProvider("doc")
	u := NewUpdater(provider, outputDir)

	err := u.Update(sourceDir, []string{"missing.go"})
	if err == nil {
		t.Fatal("expected error for missing source file")
	}
	if !strings.Contains(err.Error(), "read file") {
		t.Errorf("error = %q, want to contain 'read file'", err.Error())
	}
}

func TestUpdateProviderError(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider()
	provider.SetErrors(errors.New("ai failure"))
	u := NewUpdater(provider, outputDir)

	err := u.Update(sourceDir, []string{"main.go"})
	if err == nil {
		t.Fatal("expected error from provider")
	}
	if !strings.Contains(err.Error(), "ai failure") {
		t.Errorf("error = %q, want to contain 'ai failure'", err.Error())
	}
}

func TestUpdateCreatesNestedOutputDirs(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	nestedDir := filepath.Join(sourceDir, "pkg", "utils")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "helper.go"), []byte("package utils\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# Updated Helper")
	u := NewUpdater(provider, outputDir)

	if err := u.Update(sourceDir, []string{"pkg/utils/helper.go"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	outputFile := filepath.Join(outputDir, "pkg", "utils", "helper.go.md")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("nested output file was not created at %s", outputFile)
	}
}

func TestUpdatePromptContainsFileContent(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	content := "package main\n\nfunc Updated() {}\n"
	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# Doc")
	u := NewUpdater(provider, outputDir)

	if err := u.Update(sourceDir, []string{"main.go"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	prompts := provider.Prompts()
	if len(prompts) != 1 {
		t.Fatalf("got %d prompts, want 1", len(prompts))
	}
	if !strings.Contains(prompts[0], content) {
		t.Error("prompt does not contain file content")
	}
	if !strings.Contains(prompts[0], "updating") {
		t.Error("prompt missing update instruction")
	}
}

func TestUpdateStopsOnFirstError(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "a.go"), []byte("package a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "b.go"), []byte("package b\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider()
	provider.SetErrors(errors.New("first file fails"))
	u := NewUpdater(provider, outputDir)

	err := u.Update(sourceDir, []string{"a.go", "b.go"})
	if err == nil {
		t.Fatal("expected error")
	}

	// Should have stopped after first file
	if provider.CallCount() != 1 {
		t.Errorf("provider called %d times, want 1", provider.CallCount())
	}
}

func TestUpdateFileWithMarkdownExtension(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "docs")

	if err := os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# README\n"), 0644); err != nil {
		t.Fatal(err)
	}

	provider := mocks.NewMockProvider("# Updated README")
	u := NewUpdater(provider, outputDir)

	if err := u.Update(sourceDir, []string{"README.md"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	outputFile := filepath.Join(outputDir, "README.md.md")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("output file for markdown was not created")
	}
}
