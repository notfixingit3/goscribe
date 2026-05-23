package cmd

import (
	"fmt"

	"github.com/house/goscribe/internal/ai"
	"github.com/house/goscribe/internal/config"
	"github.com/house/goscribe/internal/docs"
	"github.com/house/goscribe/internal/git"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [path]",
	Short: "Update documentation based on git changes",
	Long: `Update existing documentation by analyzing git changes since the last 
documentation generation. Only updates sections affected by code changes.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	sourcePath := "."
	if len(args) > 0 {
		sourcePath = args[0]
	}
	
	repo, err := git.OpenRepo(sourcePath)
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}
	
	lastCommit, err := docs.LoadCommitState(sourcePath)
	if err != nil {
		return fmt.Errorf("no previous documentation state found. Run 'generate' first: %w", err)
	}
	
	currentCommit, err := repo.GetCurrentCommit()
	if err != nil {
		return fmt.Errorf("get current commit: %w", err)
	}
	
	if lastCommit == currentCommit {
		fmt.Println("No changes detected since last documentation generation.")
		return nil
	}
	
	changedFiles, err := repo.GetChangedFiles(lastCommit, currentCommit)
	if err != nil {
		return fmt.Errorf("get changed files: %w", err)
	}
	
	if len(changedFiles) == 0 {
		fmt.Println("No relevant files changed since last documentation generation.")
		return nil
	}
	
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	
	provider, err := ai.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
	}
	
	updater := docs.NewUpdater(provider, cfg.OutputDir)
	if err := updater.Update(sourcePath, changedFiles); err != nil {
		return fmt.Errorf("update docs: %w", err)
	}
	
	if err := docs.SaveCommitState(sourcePath, currentCommit); err != nil {
		return fmt.Errorf("save commit state: %w", err)
	}
	
	fmt.Printf("Documentation updated successfully (%d files changed)\n", len(changedFiles))
	return nil
}
