package docs

import (
	"strings"
	"testing"
)

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

// ---------------------------------------------------------------------------
// Coherence bridge tests
// ---------------------------------------------------------------------------

func TestCoherenceHint_WithProfileAndTemplate(t *testing.T) {
	profile := DefaultProfile()
	if profile == nil {
		t.Fatal("DefaultProfile() returned nil — was software-documenter registered?")
	}
	tmpl, ok := GetTemplate("elegant")
	if !ok {
		t.Fatal("elegant template not found")
	}
	b := NewPromptBuilder(profile).WithTemplate(tmpl)

	hint := b.coherenceHint()
	if hint == "" {
		t.Fatal("expected non-empty coherence hint")
	}
	if !strings.Contains(hint, profile.Description()) {
		t.Errorf("coherence hint should contain profile description %q, got %q", profile.Description(), hint)
	}
	if !strings.Contains(hint, tmpl.Description()) {
		t.Errorf("coherence hint should contain template description %q, got %q", tmpl.Description(), hint)
	}
}

func TestCoherenceHint_NoTemplate(t *testing.T) {
	profile := fakeProfile{name: "test"}
	b := NewPromptBuilder(profile)

	hint := b.coherenceHint()
	if hint != "" {
		t.Errorf("expected empty coherence hint, got %q", hint)
	}
}

func TestBuildGeneratePrompt_WithTemplate_IncludesCoherence(t *testing.T) {
	p := fakeProfile{name: "gen", generatePrompt: "GEN"}
	tmpl, ok := GetTemplate("elegant")
	if !ok {
		t.Fatal("elegant template not found")
	}
	b := NewPromptBuilder(p).WithTemplate(tmpl)

	got := b.BuildGeneratePrompt("main.go", []byte("package main"))
	if !strings.Contains(got, "You are generating") {
		t.Errorf("generate prompt should contain coherence bridge when template is set, got:\n%s", got)
	}
	if !strings.Contains(got, p.Description()) {
		t.Errorf("generate prompt should contain profile description, got:\n%s", got)
	}
	if !strings.Contains(got, tmpl.Description()) {
		t.Errorf("generate prompt should contain template description, got:\n%s", got)
	}
}

func TestBuildGeneratePrompt_NoTemplate_NoCoherence(t *testing.T) {
	p := fakeProfile{name: "gen", generatePrompt: "GEN"}
	b := NewPromptBuilder(p)

	got := b.BuildGeneratePrompt("main.go", []byte("package main"))
	if strings.Contains(got, "You are generating") {
		t.Errorf("generate prompt should not contain coherence bridge when no template, got:\n%s", got)
	}
	want := "GEN:main.go:package main"
	if got != want {
		t.Errorf("BuildGeneratePrompt = %q, want %q", got, want)
	}
}

func TestBuildUpdatePrompt_WithTemplate_IncludesCoherence(t *testing.T) {
	p := fakeProfile{name: "upd", updatePrompt: "UPD"}
	tmpl, ok := GetTemplate("elegant")
	if !ok {
		t.Fatal("elegant template not found")
	}
	b := NewPromptBuilder(p).WithTemplate(tmpl)

	got := b.BuildUpdatePrompt("main.go", []byte("package main"))
	if !strings.Contains(got, "You are generating") {
		t.Errorf("update prompt should contain coherence bridge when template is set, got:\n%s", got)
	}
	if !strings.Contains(got, p.Description()) {
		t.Errorf("update prompt should contain profile description, got:\n%s", got)
	}
	if !strings.Contains(got, tmpl.Description()) {
		t.Errorf("update prompt should contain template description, got:\n%s", got)
	}
}

func TestBuildUpdatePrompt_NoTemplate_NoCoherence(t *testing.T) {
	p := fakeProfile{name: "upd", updatePrompt: "UPD"}
	b := NewPromptBuilder(p)

	got := b.BuildUpdatePrompt("main.go", []byte("package main"))
	if strings.Contains(got, "You are generating") {
		t.Errorf("update prompt should not contain coherence bridge when no template, got:\n%s", got)
	}
	want := "UPD:main.go:package main"
	if got != want {
		t.Errorf("BuildUpdatePrompt = %q, want %q", got, want)
	}
}
