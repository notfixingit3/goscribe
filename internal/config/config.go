package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Provider  string `mapstructure:"provider"`
	Model     string `mapstructure:"model"`
	OutputDir string `mapstructure:"output"`
	Verbose   bool   `mapstructure:"verbose"`
}

func Load() (*Config, error) {
	viper.SetDefault("output", "docs")
	viper.SetDefault("verbose", false)

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

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
