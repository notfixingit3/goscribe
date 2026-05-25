package providers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// setupTest resets viper state and configures a temp config file.
// Returns the temp dir for cleanup.
func setupTest(t *testing.T) string {
	t.Helper()
	viper.Reset()

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, ".goscribe.yaml")
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yaml")
	// Write an empty config so viper has a file to read/write
	if err := os.WriteFile(cfgFile, []byte("providers: []\n"), 0o644); err != nil {
		t.Fatalf("setup: write config: %v", err)
	}
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("setup: read config: %v", err)
	}
	return tmpDir
}

func TestListProviders_Empty(t *testing.T) {
	setupTest(t)

	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error: %v", err)
	}
	if len(providers) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(providers))
	}
}

func TestSaveAndListProviders(t *testing.T) {
	setupTest(t)

	p := ProviderConfig{
		Name:    "ollama",
		URL:     "http://localhost:11434",
		Model:   "llama2",
		Default: true,
	}
	if err := SaveProvider(p); err != nil {
		t.Fatalf("SaveProvider() error: %v", err)
	}

	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	got := providers[0]
	if got.Name != "ollama" {
		t.Errorf("Name = %q, want %q", got.Name, "ollama")
	}
	if got.URL != "http://localhost:11434" {
		t.Errorf("URL = %q, want %q", got.URL, "http://localhost:11434")
	}
	if got.Model != "llama2" {
		t.Errorf("Model = %q, want %q", got.Model, "llama2")
	}
	if got.Default != true {
		t.Errorf("Default = %v, want true", got.Default)
	}
}

func TestSaveProvider_UpdatesExisting(t *testing.T) {
	setupTest(t)

	p1 := ProviderConfig{
		Name:  "ollama",
		URL:   "http://localhost:11434",
		Model: "llama2",
	}
	if err := SaveProvider(p1); err != nil {
		t.Fatalf("SaveProvider(p1) error: %v", err)
	}

	// Update with new model
	p2 := ProviderConfig{
		Name:  "ollama",
		URL:   "http://localhost:11434",
		Model: "llama3",
	}
	if err := SaveProvider(p2); err != nil {
		t.Fatalf("SaveProvider(p2) error: %v", err)
	}

	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider after update, got %d", len(providers))
	}
	if providers[0].Model != "llama3" {
		t.Errorf("Model = %q, want %q", providers[0].Model, "llama3")
	}
}

func TestSaveProvider_DefaultClearsOthers(t *testing.T) {
	setupTest(t)

	p1 := ProviderConfig{
		Name:    "ollama",
		URL:     "http://localhost:11434",
		Model:   "llama2",
		Default: true,
	}
	if err := SaveProvider(p1); err != nil {
		t.Fatalf("SaveProvider(p1) error: %v", err)
	}

	p2 := ProviderConfig{
		Name:    "openai",
		APIKey:  "sk-test",
		Model:   "gpt-4",
		Default: true,
	}
	if err := SaveProvider(p2); err != nil {
		t.Fatalf("SaveProvider(p2) error: %v", err)
	}

	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	// Find each and check default
	for _, p := range providers {
		if p.Name == "ollama" && p.Default {
			t.Error("ollama should not be default after openai set as default")
		}
		if p.Name == "openai" && !p.Default {
			t.Error("openai should be default")
		}
	}
}

func TestSaveProvider_MultipleProviders(t *testing.T) {
	setupTest(t)

	provs := []ProviderConfig{
		{Name: "ollama", URL: "http://localhost:11434", Model: "llama2"},
		{Name: "openai", APIKey: "sk-test", Model: "gpt-4"},
	}
	for _, p := range provs {
		if err := SaveProvider(p); err != nil {
			t.Fatalf("SaveProvider(%s) error: %v", p.Name, err)
		}
	}

	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}
}

func TestRemoveProvider(t *testing.T) {
	setupTest(t)

	p := ProviderConfig{
		Name:  "ollama",
		URL:   "http://localhost:11434",
		Model: "llama2",
	}
	if err := SaveProvider(p); err != nil {
		t.Fatalf("SaveProvider() error: %v", err)
	}

	if err := RemoveProvider("ollama"); err != nil {
		t.Fatalf("RemoveProvider() error: %v", err)
	}

	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error: %v", err)
	}
	if len(providers) != 0 {
		t.Fatalf("expected 0 providers after remove, got %d", len(providers))
	}
}

func TestRemoveProvider_NonExistent(t *testing.T) {
	setupTest(t)

	p := ProviderConfig{
		Name:  "ollama",
		URL:   "http://localhost:11434",
		Model: "llama2",
	}
	if err := SaveProvider(p); err != nil {
		t.Fatalf("SaveProvider() error: %v", err)
	}

	// Removing non-existent provider should not error, just no-op
	if err := RemoveProvider("nonexistent"); err != nil {
		t.Fatalf("RemoveProvider(nonexistent) error: %v", err)
	}

	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider still, got %d", len(providers))
	}
}

func TestRemoveProvider_OneOfMultiple(t *testing.T) {
	setupTest(t)

	p1 := ProviderConfig{Name: "ollama", URL: "http://localhost:11434", Model: "llama2"}
	p2 := ProviderConfig{Name: "openai", APIKey: "sk-test", Model: "gpt-4"}
	if err := SaveProvider(p1); err != nil {
		t.Fatalf("SaveProvider(ollama) error: %v", err)
	}
	if err := SaveProvider(p2); err != nil {
		t.Fatalf("SaveProvider(openai) error: %v", err)
	}

	if err := RemoveProvider("ollama"); err != nil {
		t.Fatalf("RemoveProvider(ollama) error: %v", err)
	}

	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	if providers[0].Name != "openai" {
		t.Errorf("remaining provider = %q, want %q", providers[0].Name, "openai")
	}
}

func TestSaveProvider_PreservesAPIKey(t *testing.T) {
	setupTest(t)

	p := ProviderConfig{
		Name:   "openai",
		APIKey: "sk-secret123",
		Model:  "gpt-4",
	}
	if err := SaveProvider(p); err != nil {
		t.Fatalf("SaveProvider() error: %v", err)
	}

	providers, err := ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error: %v", err)
	}
	if providers[0].APIKey != "sk-secret123" {
		t.Errorf("APIKey = %q, want %q", providers[0].APIKey, "sk-secret123")
	}
}

func TestProviderConfig_Fields(t *testing.T) {
	p := ProviderConfig{
		Name:    "test",
		URL:     "http://example.com",
		APIKey:  "key123",
		Model:   "model-x",
		Default: true,
	}
	if p.Name != "test" {
		t.Errorf("Name = %q, want %q", p.Name, "test")
	}
	if p.URL != "http://example.com" {
		t.Errorf("URL = %q, want %q", p.URL, "http://example.com")
	}
	if p.APIKey != "key123" {
		t.Errorf("APIKey = %q, want %q", p.APIKey, "key123")
	}
	if p.Model != "model-x" {
		t.Errorf("Model = %q, want %q", p.Model, "model-x")
	}
	if p.Default != true {
		t.Errorf("Default = %v, want true", p.Default)
	}
}
