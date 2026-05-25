package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/house/goscribe/cmd/wizard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// ExitCodeError wraps an exit code for CI mode signaling.
type ExitCodeError struct {
	Code int
}

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}

func errExit(code int) *ExitCodeError {
	return &ExitCodeError{Code: code}
}

var rootCmd = &cobra.Command{
	Use:   "goscribe",
	Short: "Generate comprehensive documentation from source code using AI",
	Long: `GoScribe reads application source code and produces complete, extensive 
user documentation with real examples. It supports OpenAI and Ollama providers 
and can track git changes for documentation updates.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("help") {
			return cmd.Help()
		}

		if viper.GetBool("ci") {
			return cmd.Help()
		}

		result, err := wizard.RunWizard()
		if err != nil {
			if errors.Is(err, wizard.ErrCancelled) {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				return nil
			}
			if errors.Is(err, wizard.ErrNoTTY) {
				return cmd.Help()
			}
			return err
		}

		if result.Profile != "" && result.Profile != "<none>" {
			viper.Set("profile", result.Profile)
		}
		if result.Template != "" && result.Template != "<none>" {
			viper.Set("template", result.Template)
		}
		viper.Set("output", result.OutputDir)

		switch result.Operation {
		case "Generate documentation":
			return runGenerate(cmd, []string{result.SourcePath})
		case "Update documentation":
			return runUpdate(cmd, []string{result.SourcePath})
		default:
			return fmt.Errorf("unknown operation: %s", result.Operation)
		}
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goscribe.yaml)")
	rootCmd.PersistentFlags().StringP("provider", "p", "", "AI provider to use (ollama, openai)")
	rootCmd.PersistentFlags().StringP("model", "m", "", "Model to use for generation")
	rootCmd.PersistentFlags().StringP("output", "o", "docs", "Output directory for documentation")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().Duration("timeout", 0, "Timeout for AI calls (e.g. 30s, 5m, default 5m0s)")
	rootCmd.PersistentFlags().Int("retries", 0, "Number of retries for transient AI errors (default 3)")
	rootCmd.PersistentFlags().Duration("retry-backoff", 0, "Initial backoff between retries (default 2s)")
	rootCmd.PersistentFlags().Bool("ci", false, "Enable CI mode (non-interactive, structured output)")
	rootCmd.PersistentFlags().String("output-format", "text", "Output format: text, json, markdown, github")
	rootCmd.PersistentFlags().String("profile", "", "Documentation profile: software-documenter, technical-writer, github-readme-expert, github-wiki-expert, api-reference, developer-onboarding, operations-runbook, release-notes, architecture-overview, contributing-guide, package-reference")
	rootCmd.PersistentFlags().String("template", "", "Documentation template: elegant, technical, futuristic, minimal, friendly")

	_ = viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider"))
	_ = viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	_ = viper.BindPFlag("retries", rootCmd.PersistentFlags().Lookup("retries"))
	_ = viper.BindPFlag("retry_backoff", rootCmd.PersistentFlags().Lookup("retry-backoff"))
	_ = viper.BindPFlag("ci", rootCmd.PersistentFlags().Lookup("ci"))
	_ = viper.BindPFlag("output_format", rootCmd.PersistentFlags().Lookup("output-format"))
	_ = viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	_ = viper.BindPFlag("template", rootCmd.PersistentFlags().Lookup("template"))
}

func initConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir: %w", err)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".goscribe")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("GOSCRIBE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config: %w", err)
		}
	}

	return nil
}
