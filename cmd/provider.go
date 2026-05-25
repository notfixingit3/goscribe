package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/house/goscribe/pkg/providers"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage AI providers",
	Long:  `Add, list, and configure AI providers (OpenAI, Ollama).`,
}

var providerAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new AI provider",
	Long: `Add a new AI provider configuration.

Examples:
  goscribe provider add ollama --url http://localhost:11434 --model llama2
  goscribe provider add openai --key sk-... --model gpt-4`,
	Args: cobra.ExactArgs(1),
	RunE: runProviderAdd,
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured providers",
	RunE:  runProviderList,
}

var providerRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a provider configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runProviderRemove,
}

func init() {
	providerAddCmd.Flags().StringP("url", "u", "", "Provider URL (for Ollama)")
	providerAddCmd.Flags().StringP("key", "k", "", "API key (for OpenAI)")
	providerAddCmd.Flags().StringP("model", "m", "", "Default model")
	providerAddCmd.Flags().Bool("default", false, "Set as default provider")

	providerCmd.AddCommand(providerAddCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerRemoveCmd)
	rootCmd.AddCommand(providerCmd)
}

func runProviderAdd(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(strings.TrimSpace(args[0]))

	if name == "" {
		return fmt.Errorf("provider name cannot be empty.\nUsage: goscribe provider add <name>\nSupported providers: openai, ollama")
	}

	supportedProviders := map[string]bool{"openai": true, "ollama": true}
	if !supportedProviders[name] {
		return fmt.Errorf("unsupported provider %q.\nSupported providers: openai, ollama", name)
	}

	urlVal, _ := cmd.Flags().GetString("url")
	key, _ := cmd.Flags().GetString("key")
	model, _ := cmd.Flags().GetString("model")
	isDefault, _ := cmd.Flags().GetBool("default")

	if model == "" {
		return fmt.Errorf("--model is required.\nUsage:\n  goscribe provider add %s --model <model> %s\nExample models: gpt-4, llama2, mistral", name, providerFlagsHint(name))
	}

	// Validate provider-specific requirements
	if name == "openai" {
		if key == "" {
			return fmt.Errorf("OpenAI requires an API key (--key).\nGet your key at https://platform.openai.com/api-keys\nUsage: goscribe provider add openai --key <KEY> --model %s", model)
		}
		if !strings.HasPrefix(key, "sk-") {
			return fmt.Errorf("OpenAI API key should start with 'sk-'. Got key starting with %q.\nCheck your key at https://platform.openai.com/api-keys", key[:min(10, len(key))])
		}
	}

	if name == "ollama" {
		if urlVal != "" {
			parsed, err := url.Parse(urlVal)
			if err != nil {
				return fmt.Errorf("invalid URL %q: %w.\nUse format: http://host:port (e.g. http://localhost:11434)", urlVal, err)
			}
			if parsed.Scheme == "" {
				return fmt.Errorf("URL %q is missing a scheme.\nUse format: http://host:port (e.g. http://localhost:11434)", urlVal)
			}
		}
	}

	provider := providers.ProviderConfig{
		Name:    name,
		URL:     urlVal,
		APIKey:  key,
		Model:   model,
		Default: isDefault,
	}

	if err := providers.SaveProvider(provider); err != nil {
		return fmt.Errorf("save provider: %w.\nCheck that ~/.goscribe.yaml is writable", err)
	}

	fmt.Printf("Provider '%s' added successfully (model: %s)\n", name, model)
	return nil
}

func runProviderList(cmd *cobra.Command, args []string) error {
	providerList, err := providers.ListProviders()
	if err != nil {
		return fmt.Errorf("list providers: %w", err)
	}

	if len(providerList) == 0 {
		fmt.Println("No providers configured.")
		fmt.Println("Add one with:")
		fmt.Println("  goscribe provider add ollama --url http://localhost:11434 --model llama2")
		fmt.Println("  goscribe provider add openai --key <API_KEY> --model gpt-4")
		return nil
	}

	fmt.Println("Configured providers:")
	for _, p := range providerList {
		defaultMark := ""
		if p.Default {
			defaultMark = " (default)"
		}
		fmt.Printf("  - %s%s\n", p.Name, defaultMark)
		if p.URL != "" {
			fmt.Printf("    URL: %s\n", p.URL)
		}
		if p.Model != "" {
			fmt.Printf("    Model: %s\n", p.Model)
		}
	}

	return nil
}

func runProviderRemove(cmd *cobra.Command, args []string) error {
	name := strings.TrimSpace(args[0])

	if name == "" {
		return fmt.Errorf("provider name cannot be empty.\nUsage: goscribe provider remove <name>")
	}

	// Check if provider exists before removing
	providerList, err := providers.ListProviders()
	if err != nil {
		return fmt.Errorf("list providers: %w", err)
	}

	found := false
	for _, p := range providerList {
		if p.Name == name {
			found = true
			break
		}
	}
	if !found {
		if len(providerList) == 0 {
			return fmt.Errorf("provider %q not found. No providers are currently configured", name)
		}
		names := make([]string, len(providerList))
		for i, p := range providerList {
			names[i] = p.Name
		}
		return fmt.Errorf("provider %q not found. Configured providers: %s", name, names)
	}

	if err := providers.RemoveProvider(name); err != nil {
		return fmt.Errorf("remove provider: %w", err)
	}

	fmt.Printf("Provider '%s' removed successfully\n", name)
	return nil
}

func providerFlagsHint(name string) string {
	switch name {
	case "openai":
		return "--key <API_KEY>"
	case "ollama":
		return "--url http://localhost:11434"
	default:
		return ""
	}
}
