package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/house/goscribe/internal/ai"
	"github.com/house/goscribe/internal/config"
	"github.com/house/goscribe/internal/docs"
	"github.com/house/goscribe/internal/git"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateCmd = &cobra.Command{
	Use:   "generate [path]",
	Short: "Generate documentation from source code",
	Long: `Generate comprehensive documentation from application source code.
Reads the source directory and produces detailed documentation with real examples.

If the path is a git repository, GoScribe will track the current commit hash
for future incremental updates.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().BoolP("force", "f", false, "Force regeneration even if docs exist")
	viper.BindPFlag("force", generateCmd.Flags().Lookup("force"))
	
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	sourcePath := "."
	if len(args) > 0 {
		sourcePath = args[0]
	}
	
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}
	
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	
	provider, err := ai.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
	}
	
	generator := docs.NewGenerator(provider, cfg.OutputDir)
	
	if err := generator.Generate(absPath); err != nil {
		return fmt.Errorf("generate docs: %w", err)
	}
	
	repo, err := git.OpenRepo(absPath)
	if err == nil {
		commit, err := repo.GetCurrentCommit()
		if err == nil {
			if err := docs.SaveCommitState(absPath, commit); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save commit state: %v\n", err)
			}
		}
	}
	
	fmt.Printf("Documentation generated successfully in %s\n", cfg.OutputDir)
	return nil
}
