// Package opencode provides the OpenCode plugin client for integrating GoScribe
// documentation generation into IDE and editor environments.
package opencode

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PluginConfig holds the configuration for the OpenCode plugin, including provider
// settings, file watch patterns, and agent connection details.
type PluginConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AutoGenerate   bool     `yaml:"auto_generate"`
	OutputDir      string   `yaml:"output_dir"`
	ProviderName   string   `yaml:"provider"`
	Model          string   `yaml:"model"`
	ProjectPath    string   `yaml:"project_path"`
	WatchPatterns  []string `yaml:"watch_patterns"`
	IgnorePatterns []string `yaml:"ignore_patterns"`
	AgentAddress   string   `yaml:"agent_address"`
	Profile        string   `yaml:"profile"`
}

// DefaultConfig returns a PluginConfig with sensible defaults.
func DefaultConfig() PluginConfig {
	return PluginConfig{
		Enabled:        true,
		AutoGenerate:   false,
		OutputDir:      "docs",
		WatchPatterns:  []string{"**/*.go"},
		IgnorePatterns: []string{"**/*_test.go", "vendor/**"},
		AgentAddress:   "http://localhost:8080",
	}
}

// LoadConfig reads and parses a YAML configuration file at the given path.
// Returns the default config if the file does not exist.
func LoadConfig(path string) (PluginConfig, error) {
	cfg := DefaultConfig()

	path = filepath.Clean(path)
	data, err := os.ReadFile(path) // #nosec G304 -- path is cleaned before read
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// LoadConfigFromProject loads plugin configuration from the standard project
// location: <projectPath>/.opencode/goscribe.yaml.
func LoadConfigFromProject(projectPath string) (PluginConfig, error) {
	path := filepath.Join(projectPath, ".opencode", "goscribe.yaml")
	return LoadConfig(path)
}

// Save writes the configuration as YAML to the given file path, creating
// parent directories as needed.
func (c *PluginConfig) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
