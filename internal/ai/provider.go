package ai

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/house/goscribe/internal/config"
	"github.com/house/goscribe/pkg/providers"
)

// Provider is the interface for AI text generation backends.
type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// NewProvider creates an AI provider based on the given configuration.
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
		if selected == nil {
			names := providerNames(providerList)
			if len(names) == 0 {
				return nil, fmt.Errorf("provider %q not found and no providers are configured.\nAdd a provider with:\n  goscribe provider add %s --url <URL> --model <MODEL>", cfg.Provider, cfg.Provider)
			}
			return nil, fmt.Errorf("provider %q not found. Available providers: %s\nAdd it with: goscribe provider add %s --url <URL> --model <MODEL>", cfg.Provider, names, cfg.Provider)
		}
	} else {
		for _, p := range providerList {
			if p.Default {
				selected = &p
				break
			}
		}
		if selected == nil && len(providerList) > 0 {
			selected = &providerList[0]
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("no providers configured.\nAdd one with:\n  goscribe provider add ollama --url http://localhost:11434 --model llama2\n  goscribe provider add openai --key <API_KEY> --model gpt-4")
	}

	factory, ok := defaultRegistry.Get(selected.Name)
	if !ok {
		registered := RegisteredProviders()
		return nil, fmt.Errorf("unsupported provider %q. Registered providers: %s.\nRemove it with: goscribe provider remove %s", selected.Name, registered, selected.Name)
	}

	inner, err := factory(*selected)
	if err != nil {
		return nil, err
	}

	if oai, ok := inner.(*OpenAIClient); ok && cfg.Timeout > 0 {
		oai.SetTimeout(cfg.Timeout)
	}
	if cc, ok := inner.(*ClaudeClient); ok && cfg.Timeout > 0 {
		cc.SetTimeout(cfg.Timeout)
	}
	if gc, ok := inner.(*GeminiClient); ok && cfg.Timeout > 0 {
		gc.SetTimeout(cfg.Timeout)
	}

	// Wrap with retry if configured
	retries := cfg.Retries
	if retries < 0 {
		retries = 0
	}
	if retries > 0 {
		verbose := func(msg string) {}
		if cfg.Verbose {
			verbose = func(msg string) {
				fmt.Fprintf(os.Stderr, "[retry] %s\n", msg)
			}
		}
		return NewRetryProvider(inner, retries, cfg.RetryBackoff, WithVerbose(verbose), withWriter(io.Discard)), nil
	}

	return inner, nil
}

func providerNames(list []providers.ProviderConfig) string {
	if len(list) == 0 {
		return "(none)"
	}
	names := make([]string, len(list))
	for i, p := range list {
		names[i] = p.Name
	}
	return fmt.Sprintf("%v", names)
}

// withWriter is a placeholder option — kept for future log-to-file support.
func withWriter(w io.Writer) RetryOption {
	return func(r *RetryProvider) {}
}

func openaiFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI provider requires an API key.\nSet it with:\n  goscribe provider add openai --key <YOUR_API_KEY> --model %s", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://api.openai.com"
	}
	return NewOpenAIClient(cfg.APIKey, cfg.Model, 0), nil
}

func ollamaFactory(cfg providers.ProviderConfig) (Provider, error) {
	return NewOllamaClient(cfg.URL, cfg.Model)
}

func kimiFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("kimi provider requires an API key; set it with: goscribe provider add kimi --key <YOUR_API_KEY> --model %s (get a key at https://platform.moonshot.cn)", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://api.moonshot.cn"
	}
	return NewKimiClient(cfg.APIKey, cfg.Model, 0), nil
}

func openrouterFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openrouter provider requires an API key; set it with: goscribe provider add openrouter --key <YOUR_API_KEY> --model %s or set OPENROUTER_API_KEY", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://openrouter.ai/api"
	}
	return NewOpenRouterClient(cfg.APIKey, cfg.Model, 0), nil
}

func claudeFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("claude provider requires an API key; set it with: goscribe provider add claude --key <YOUR_API_KEY> --model %s (get a key at https://console.anthropic.com)", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://api.anthropic.com"
	}
	return NewClaudeClient(cfg.APIKey, cfg.Model, 0), nil
}

func geminiFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini provider requires an API key; set it with: goscribe provider add gemini --key <YOUR_API_KEY> --model %s or set GOOGLE_API_KEY (get a key at https://aistudio.google.com/apikey)", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://generativelanguage.googleapis.com"
	}
	return NewGeminiClient(cfg.APIKey, cfg.Model, 0), nil
}

func lmstudioFactory(cfg providers.ProviderConfig) (Provider, error) {
	return NewLMStudioClient(cfg.URL, cfg.Model), nil
}

func llamacppFactory(cfg providers.ProviderConfig) (Provider, error) {
	return NewLLaMACppClient(cfg.URL, cfg.Model), nil
}

func ollamaCloudFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("ollama-cloud provider requires an API key; set it with: goscribe provider add ollama-cloud --url <URL> --key <YOUR_API_KEY> --model %s", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://cloud.ollama.com"
	}
	return NewOllamaCloudClient(cfg.APIKey, cfg.URL, cfg.Model)
}

func ollamaSelfHostedFactory(cfg providers.ProviderConfig) (Provider, error) {
	return NewOllamaSelfHostedClient(cfg.URL, cfg.Model)
}

func opencodeFactory(cfg providers.ProviderConfig) (Provider, error) {
	return NewOpenCodeClient(cfg.URL, cfg.Model), nil
}

func xaiFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("xai provider requires an API key; set it with: goscribe provider add xai --key <YOUR_API_KEY> --model %s or set XAI_API_KEY", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://api.x.ai"
	}
	return NewXAIClient(cfg.APIKey, cfg.URL, cfg.Model), nil
}

func nvidiaFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("nvidia provider requires an API key; set it with: goscribe provider add nvidia --key <YOUR_API_KEY> --model %s or set NVIDIA_API_KEY", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://integrate.api.nvidia.com"
	}
	return NewNVIDIAClient(cfg.APIKey, cfg.URL, cfg.Model), nil
}

func githubCopilotFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("github-copilot provider requires a token; set it with: goscribe provider add github-copilot --key <YOUR_TOKEN> --model %s", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://api.githubcopilot.com"
	}
	return NewGitHubCopilotClient(cfg.APIKey, cfg.URL, cfg.Model), nil
}

func zaiFactory(cfg providers.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("zai provider requires an API key; set it with: goscribe provider add zai --key <YOUR_API_KEY> --model %s or set ZAI_API_KEY", cfg.Model)
	}
	if cfg.URL == "" {
		cfg.URL = "https://api.z.ai"
	}
	return NewZaiClient(cfg.APIKey, cfg.URL, cfg.Model), nil
}

func vllmFactory(cfg providers.ProviderConfig) (Provider, error) {
	return NewVLLMClient(cfg.URL, cfg.Model), nil
}

func init() {
	RegisterProvider("openai", openaiFactory)
	RegisterProvider("ollama", ollamaFactory)
	RegisterProvider("kimi", kimiFactory)
	RegisterProvider("openrouter", openrouterFactory)
	RegisterProvider("claude", claudeFactory)
	RegisterProvider("gemini", geminiFactory)
	RegisterProvider("lmstudio", lmstudioFactory)
	RegisterProvider("llamacpp", llamacppFactory)
	RegisterProvider("ollama-cloud", ollamaCloudFactory)
	RegisterProvider("ollama-selfhosted", ollamaSelfHostedFactory)
	RegisterProvider("opencode", opencodeFactory)
	RegisterProvider("xai", xaiFactory)
	RegisterProvider("nvidia", nvidiaFactory)
	RegisterProvider("github-copilot", githubCopilotFactory)
	RegisterProvider("zai", zaiFactory)
	RegisterProvider("vllm", vllmFactory)
}
