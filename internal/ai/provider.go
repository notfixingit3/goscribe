package ai

import (
	"context"
	"fmt"

	"github.com/house/goscribe/internal/config"
	"github.com/house/goscribe/pkg/providers"
)

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

func NewProvider(cfg *config.Config) (Provider, error) {
	providerList, err := providers.ListProviders()
	if err != nil {
		return nil, fmt.Errorf("list providers: %w", err)
	}

	var selected *providers.ProviderConfig
	
	if cfg.Provider != "" {
		for _, p := range providerList {
			if p.Name == cfg.Provider {
				selected = &p
				break
			}
		}
	} else {
		for _, p := range providerList {
			if p.Default {
				selected = &p
				break
			}
		}
	}

	if selected == nil {
		if len(providerList) > 0 {
			selected = &providerList[0]
		} else {
			return nil, fmt.Errorf("no providers configured. Use 'goscribe provider add' to add one")
		}
	}

	switch selected.Name {
	case "openai":
		return NewOpenAIClient(selected.APIKey, selected.Model), nil
	case "ollama":
		return NewOllamaClient(selected.URL, selected.Model)
	default:
		return nil, fmt.Errorf("unknown provider: %s", selected.Name)
	}
}
