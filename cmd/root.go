package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

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
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goscribe.yaml)")
	rootCmd.PersistentFlags().StringP("provider", "p", "", "AI provider to use (ollama, openai)")
	rootCmd.PersistentFlags().StringP("model", "m", "", "Model to use for generation")
	rootCmd.PersistentFlags().StringP("output", "o", "docs", "Output directory for documentation")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	
	viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider"))
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
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
