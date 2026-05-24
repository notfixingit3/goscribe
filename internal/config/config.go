// Package config handles loading and saving goscribe configuration.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds all goscribe configuration values.
type Config struct {
	Provider     string        `mapstructure:"provider"`
	Model        string        `mapstructure:"model"`
	OutputDir    string        `mapstructure:"output"`
	Verbose      bool          `mapstructure:"verbose"`
	Timeout      time.Duration `mapstructure:"timeout"`
	Retries      int           `mapstructure:"retries"`
	RetryBackoff time.Duration `mapstructure:"retry_backoff"`
	CI           bool          `mapstructure:"ci"`
	OutputFormat string        `mapstructure:"output_format"`
	Profile      string        `mapstructure:"profile"`
}

// Load reads configuration from viper (flags, env, config file) and returns a Config.
func Load() (*Config, error) {
	viper.SetDefault("output", "docs")
	viper.SetDefault("verbose", false)
	viper.SetDefault("timeout", 5*time.Minute)
	viper.SetDefault("retries", 3)
	viper.SetDefault("retry_backoff", 2*time.Second)

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to the default config file path.
func Save(cfg *Config) error {
	viper.Set("provider", cfg.Provider)
	viper.Set("model", cfg.Model)
	viper.Set("output", cfg.OutputDir)
	viper.Set("verbose", cfg.Verbose)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	configPath := home + "/.goscribe.yaml"
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
