package providers

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type ProviderConfig struct {
	Name    string `mapstructure:"name"`
	URL     string `mapstructure:"url"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	Default bool   `mapstructure:"default"`
}

func SaveProvider(provider ProviderConfig) error {
	providers, err := ListProviders()
	if err != nil {
		return fmt.Errorf("list providers: %w", err)
	}

	if provider.Default {
		for i := range providers {
			providers[i].Default = false
		}
	}

	found := false
	for i, p := range providers {
		if p.Name == provider.Name {
			providers[i] = provider
			found = true
			break
		}
	}

	if !found {
		providers = append(providers, provider)
	}

	viper.Set("providers", providers)
	return saveConfig()
}

func ListProviders() ([]ProviderConfig, error) {
	var providers []ProviderConfig
	if err := viper.UnmarshalKey("providers", &providers); err != nil {
		return nil, fmt.Errorf("unmarshal providers: %w", err)
	}
	return providers, nil
}

func RemoveProvider(name string) error {
	providers, err := ListProviders()
	if err != nil {
		return fmt.Errorf("list providers: %w", err)
	}

	var filtered []ProviderConfig
	for _, p := range providers {
		if p.Name != name {
			filtered = append(filtered, p)
		}
	}

	viper.Set("providers", filtered)
	return saveConfig()
}

func saveConfig() error {
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
