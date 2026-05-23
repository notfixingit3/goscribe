package cmd

import (
	"fmt"

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
	name := args[0]
	
	url, _ := cmd.Flags().GetString("url")
	key, _ := cmd.Flags().GetString("key")
	model, _ := cmd.Flags().GetString("model")
	isDefault, _ := cmd.Flags().GetBool("default")
	
	provider := providers.ProviderConfig{
		Name:    name,
		URL:     url,
		APIKey:  key,
		Model:   model,
		Default: isDefault,
	}
	
	if err := providers.SaveProvider(provider); err != nil {
		return fmt.Errorf("save provider: %w", err)
	}
	
	fmt.Printf("Provider '%s' added successfully\n", name)
	return nil
}

func runProviderList(cmd *cobra.Command, args []string) error {
	providerList, err := providers.ListProviders()
	if err != nil {
		return fmt.Errorf("list providers: %w", err)
	}
	
	if len(providerList) == 0 {
		fmt.Println("No providers configured.")
		fmt.Println("Use 'goscribe provider add' to add one.")
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
	name := args[0]
	
	if err := providers.RemoveProvider(name); err != nil {
		return fmt.Errorf("remove provider: %w", err)
	}
	
	fmt.Printf("Provider '%s' removed successfully\n", name)
	return nil
}
