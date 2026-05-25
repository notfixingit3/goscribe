package docs

import (
	"strings"
	"testing"
)

func TestDefaultProfile_PromptNotEmpty(t *testing.T) {
	p := softwareDocumenterProfile{}
	file := "main.go"
	content := []byte("package main\n\nfunc main() {}")

	prompt := p.BuildPrompt(file, content)
	if len(prompt) <= 100 {
		t.Fatalf("BuildPrompt output is %d chars, expected >100", len(prompt))
	}
}

func TestDefaultProfile_HasPersona(t *testing.T) {
	p := softwareDocumenterProfile{}
	file := "main.go"
	content := []byte("package main\n\nfunc main() {}")

	prompt := p.BuildPrompt(file, content)
	if !strings.Contains(prompt, "You are a") {
		t.Fatal("BuildPrompt output should contain a persona statement \"You are a\"")
	}
}

func TestDefaultProfile_HasStructure(t *testing.T) {
	p := softwareDocumenterProfile{}
	file := "main.go"
	content := []byte("package main\n\nfunc main() {}")

	prompt := p.BuildPrompt(file, content)
	if !strings.Contains(prompt, "Overview") && !strings.Contains(prompt, "Structure") {
		t.Fatal("BuildPrompt output should contain \"Overview\" or \"Structure\" guidance")
	}
}

func TestDefaultProfile_UpdatePromptUpgraded(t *testing.T) {
	p := softwareDocumenterProfile{}
	file := "main.go"
	content := []byte("package main\n\nfunc main() {}")

	prompt := p.BuildUpdatePrompt(file, content)
	if len(prompt) <= 80 {
		t.Fatalf("BuildUpdatePrompt output is %d chars, expected >80", len(prompt))
	}
}

func TestDefaultProfile_NameUnchanged(t *testing.T) {
	p := softwareDocumenterProfile{}
	if got := p.Name(); got != DefaultProfileName {
		t.Fatalf("Name() = %q, want %q", got, DefaultProfileName)
	}
}

func TestDefaultProfile_DescriptionUnchanged(t *testing.T) {
	p := softwareDocumenterProfile{}
	want := "General software documentation with examples and explanations"
	if got := p.Description(); got != want {
		t.Fatalf("Description() = %q, want %q", got, want)
	}
}

func TestDefaultProfile_FormatOutputPassthrough(t *testing.T) {
	p := softwareDocumenterProfile{}
	input := "# Some markdown\n\nContent here."
	if got := p.FormatOutput(input); got != input {
		t.Fatalf("FormatOutput() = %q, want unmodified input", got)
	}
}

func TestDefaultProfile_IncludesFileAndContent(t *testing.T) {
	p := softwareDocumenterProfile{}
	file := "handler.go"
	content := []byte("func Handle() { return }")

	prompt := p.BuildPrompt(file, content)
	if !strings.Contains(prompt, file) {
		t.Fatal("BuildPrompt output should include the file name")
	}
	if !strings.Contains(prompt, string(content)) {
		t.Fatal("BuildPrompt output should include the source content")
	}
}

func TestDefaultProfile_UpdatePromptIncludesFileAndContent(t *testing.T) {
	p := softwareDocumenterProfile{}
	file := "handler.go"
	content := []byte("func Handle() { return }")

	prompt := p.BuildUpdatePrompt(file, content)
	if !strings.Contains(prompt, file) {
		t.Fatal("BuildUpdatePrompt output should include the file name")
	}
	if !strings.Contains(prompt, string(content)) {
		t.Fatal("BuildUpdatePrompt output should include the source content")
	}
}
