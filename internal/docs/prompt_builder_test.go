package docs

import "testing"

type fakeProfile struct {
	name            string
	generatePrompt  string
	updatePrompt    string
	formattedOutput string
}

func (f fakeProfile) Name() string        { return f.name }
func (f fakeProfile) Description() string { return "fake: " + f.name }
func (f fakeProfile) BuildPrompt(file string, content []byte) string {
	return f.generatePrompt + ":" + file + ":" + string(content)
}
func (f fakeProfile) BuildUpdatePrompt(file string, content []byte) string {
	return f.updatePrompt + ":" + file + ":" + string(content)
}
func (f fakeProfile) FormatOutput(doc string) string {
	return f.formattedOutput + ":" + doc
}

func TestNewPromptBuilder_NilProfileUsesDefault(t *testing.T) {
	withIsolatedDefaultProfileRegistry(t)
	RegisterProfile(fakeProfile{name: DefaultProfileName})

	b := NewPromptBuilder(nil)
	if b.profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if b.profile.Name() != DefaultProfileName {
		t.Fatalf("expected default profile, got %q", b.profile.Name())
	}
}

func TestNewPromptBuilder_WithProfile(t *testing.T) {
	p := fakeProfile{name: "custom"}
	b := NewPromptBuilder(p)
	if b.profile.Name() != "custom" {
		t.Fatalf("expected custom profile, got %q", b.profile.Name())
	}
}

func TestPromptBuilder_BuildGeneratePrompt(t *testing.T) {
	p := fakeProfile{name: "gen", generatePrompt: "GEN"}
	b := NewPromptBuilder(p)

	got := b.BuildGeneratePrompt("main.go", []byte("package main"))
	want := "GEN:main.go:package main"
	if got != want {
		t.Fatalf("BuildGeneratePrompt = %q, want %q", got, want)
	}
}

func TestPromptBuilder_BuildUpdatePrompt(t *testing.T) {
	p := fakeProfile{name: "upd", updatePrompt: "UPD"}
	b := NewPromptBuilder(p)

	got := b.BuildUpdatePrompt("main.go", []byte("package main"))
	want := "UPD:main.go:package main"
	if got != want {
		t.Fatalf("BuildUpdatePrompt = %q, want %q", got, want)
	}
}

func TestPromptBuilder_FormatOutput(t *testing.T) {
	p := fakeProfile{name: "fmt", formattedOutput: "FMT"}
	b := NewPromptBuilder(p)

	got := b.FormatOutput("raw doc")
	want := "FMT:raw doc"
	if got != want {
		t.Fatalf("FormatOutput = %q, want %q", got, want)
	}
}
