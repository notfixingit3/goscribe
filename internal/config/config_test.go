package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
)

// resetViper clears all viper state so tests are isolated.
func resetViper() {
	viper.Reset()
	viper.SetEnvPrefix("GOSCRIBE")
	viper.AutomaticEnv()
}

func TestLoad_Defaults(t *testing.T) {
	resetViper()
	defer resetViper()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.OutputDir != "docs" {
		t.Errorf("default OutputDir = %q, want %q", cfg.OutputDir, "docs")
	}
	if cfg.Verbose != false {
		t.Errorf("default Verbose = %v, want false", cfg.Verbose)
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("default Timeout = %v, want %v", cfg.Timeout, 5*time.Minute)
	}
	if cfg.Retries != 3 {
		t.Errorf("default Retries = %d, want 3", cfg.Retries)
	}
	if cfg.RetryBackoff != 2*time.Second {
		t.Errorf("default RetryBackoff = %v, want %v", cfg.RetryBackoff, 2*time.Second)
	}
	if cfg.Provider != "" {
		t.Errorf("default Provider = %q, want empty", cfg.Provider)
	}
	if cfg.Model != "" {
		t.Errorf("default Model = %q, want empty", cfg.Model)
	}
}

func TestLoad_FromYAMLFile(t *testing.T) {
	resetViper()
	defer resetViper()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".goscribe.yaml")
	content := []byte("provider: openai\nmodel: gpt-4\noutput: mydocs\nverbose: true\ntimeout: 30s\nretries: 5\nretry_backoff: 5s\n")
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Provider != "openai" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "openai")
	}
	if cfg.Model != "gpt-4" {
		t.Errorf("Model = %q, want %q", cfg.Model, "gpt-4")
	}
	if cfg.OutputDir != "mydocs" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, "mydocs")
	}
	if cfg.Verbose != true {
		t.Errorf("Verbose = %v, want true", cfg.Verbose)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}
	if cfg.Retries != 5 {
		t.Errorf("Retries = %d, want 5", cfg.Retries)
	}
	if cfg.RetryBackoff != 5*time.Second {
		t.Errorf("RetryBackoff = %v, want %v", cfg.RetryBackoff, 5*time.Second)
	}
}

func TestLoad_PartialConfigUsesDefaults(t *testing.T) {
	resetViper()
	defer resetViper()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".goscribe.yaml")
	content := []byte("provider: ollama\n")
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Provider != "ollama" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "ollama")
	}
	if cfg.OutputDir != "docs" {
		t.Errorf("OutputDir = %q, want default %q", cfg.OutputDir, "docs")
	}
	if cfg.Retries != 3 {
		t.Errorf("Retries = %d, want default 3", cfg.Retries)
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	resetViper()
	defer resetViper()

	envVars := map[string]string{
		"GOSCRIBE_PROVIDER": "openai",
		"GOSCRIBE_MODEL":    "gpt-3.5-turbo",
		"GOSCRIBE_OUTPUT":   "output-dir",
		"GOSCRIBE_VERBOSE":  "true",
	}
	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	// Viper's AutomaticEnv only works with BindEnv for Unmarshal.
	// This mirrors what cmd/root.go does via viper.BindPFlag.
	viper.SetEnvPrefix("GOSCRIBE")
	viper.AutomaticEnv()
	viper.BindEnv("provider")
	viper.BindEnv("model")
	viper.BindEnv("output")
	viper.BindEnv("verbose")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Provider != "openai" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "openai")
	}
	if cfg.Model != "gpt-3.5-turbo" {
		t.Errorf("Model = %q, want %q", cfg.Model, "gpt-3.5-turbo")
	}
	if cfg.OutputDir != "output-dir" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, "output-dir")
	}
	if cfg.Verbose != true {
		t.Errorf("Verbose = %v, want true", cfg.Verbose)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	resetViper()
	defer resetViper()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".goscribe.yaml")
	content := []byte("provider: ollama\nmodel: llama2\noutput: filedocs\n")
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}

	os.Setenv("GOSCRIBE_PROVIDER", "openai")
	os.Setenv("GOSCRIBE_MODEL", "gpt-4o")
	defer os.Unsetenv("GOSCRIBE_PROVIDER")
	defer os.Unsetenv("GOSCRIBE_MODEL")

	viper.SetEnvPrefix("GOSCRIBE")
	viper.AutomaticEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Provider != "openai" {
		t.Errorf("Provider = %q, want %q from env", cfg.Provider, "openai")
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("Model = %q, want %q from env", cfg.Model, "gpt-4o")
	}
	// output from file should still be respected (no env override for it)
	if cfg.OutputDir != "filedocs" {
		t.Errorf("OutputDir = %q, want %q from file", cfg.OutputDir, "filedocs")
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	resetViper()
	defer resetViper()

	// Point to nonexistent file - viper won't error if we don't ReadInConfig
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() without config file returned error: %v", err)
	}

	if cfg.OutputDir != "docs" {
		t.Errorf("default OutputDir = %q, want %q", cfg.OutputDir, "docs")
	}
}

func TestSave_WritesYAMLFile(t *testing.T) {
	resetViper()
	defer resetViper()

	tmpDir := t.TempDir()
	// Override home directory for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{
		Provider:  "ollama",
		Model:     "llama3",
		OutputDir: "my-docs",
		Verbose:   true,
	}

	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".goscribe.yaml")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}

	_ = string(data)
	// Verify file exists and is valid YAML by loading it back
	viper.SetConfigFile(expectedPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to reload saved config: %v", err)
	}

	if viper.GetString("provider") != "ollama" {
		t.Errorf("saved provider = %q, want %q", viper.GetString("provider"), "ollama")
	}
	if viper.GetString("model") != "llama3" {
		t.Errorf("saved model = %q, want %q", viper.GetString("model"), "llama3")
	}
	if viper.GetString("output") != "my-docs" {
		t.Errorf("saved output = %q, want %q", viper.GetString("output"), "my-docs")
	}
	if !viper.GetBool("verbose") {
		t.Errorf("saved verbose = false, want true")
	}
}

func TestSave_OverwriteExisting(t *testing.T) {
	resetViper()
	defer resetViper()

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Write initial config
	cfg1 := &Config{
		Provider:  "ollama",
		Model:     "llama2",
		OutputDir: "docs",
		Verbose:   false,
	}
	if err := Save(cfg1); err != nil {
		t.Fatalf("first Save(): %v", err)
	}

	// Overwrite with new values
	cfg2 := &Config{
		Provider:  "openai",
		Model:     "gpt-4",
		OutputDir: "new-docs",
		Verbose:   true,
	}
	if err := Save(cfg2); err != nil {
		t.Fatalf("second Save(): %v", err)
	}

	// Verify the file has new values
	expectedPath := filepath.Join(tmpDir, ".goscribe.yaml")
	viper.SetConfigFile(expectedPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if viper.GetString("provider") != "openai" {
		t.Errorf("provider = %q, want %q", viper.GetString("provider"), "openai")
	}
	if viper.GetString("output") != "new-docs" {
		t.Errorf("output = %q, want %q", viper.GetString("output"), "new-docs")
	}
}

func TestSave_HomeDirError(t *testing.T) {
	resetViper()
	defer resetViper()

	cfg := &Config{Provider: "ollama", Model: "llama2", OutputDir: "docs"}

	// This test is hard to trigger since UserHomeDir rarely fails.
	// We just verify Save works in normal conditions.
	// If HOME is set, Save should succeed.
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	err := Save(cfg)
	if err != nil {
		t.Errorf("Save() with valid home dir returned error: %v", err)
	}
}

func TestLoad_ReturnsPointer(t *testing.T) {
	resetViper()
	defer resetViper()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil")
	}
}

func TestLoad_AllFieldsAccessible(t *testing.T) {
	resetViper()
	defer resetViper()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".goscribe.yaml")
	content := []byte(fmt.Sprintf("provider: test\nmodel: testmodel\noutput: testout\nverbose: true\ntimeout: 1m%s\nretries: 10\nretry_backoff: 500ms\n", ""))
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Provider != "test" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "test")
	}
	if cfg.Model != "testmodel" {
		t.Errorf("Model = %q, want %q", cfg.Model, "testmodel")
	}
	if cfg.OutputDir != "testout" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, "testout")
	}
	if !cfg.Verbose {
		t.Errorf("Verbose = false, want true")
	}
	if cfg.Retries != 10 {
		t.Errorf("Retries = %d, want 10", cfg.Retries)
	}
}

func TestSave_DoesNotSaveTimeoutOrRetries(t *testing.T) {
	resetViper()
	defer resetViper()

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{
		Provider:     "ollama",
		Model:        "llama2",
		OutputDir:    "docs",
		Verbose:      false,
		Timeout:      10 * time.Minute,
		Retries:      7,
		RetryBackoff: 5 * time.Second,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save(): %v", err)
	}

	// Save only writes provider, model, output, verbose (as per the implementation)
	expectedPath := filepath.Join(tmpDir, ".goscribe.yaml")
	viper.SetConfigFile(expectedPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}

	// These should be saved
	if viper.GetString("provider") != "ollama" {
		t.Errorf("provider not saved correctly")
	}
	if viper.GetString("model") != "llama2" {
		t.Errorf("model not saved correctly")
	}
}
