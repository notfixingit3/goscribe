package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/house/goscribe/internal/ai"
	"github.com/house/goscribe/internal/ci"
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

	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("resolve path %q: %w", sourcePath, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s\nCheck the path and try again", absPath)
		}
		return fmt.Errorf("cannot access path %s: %w", absPath, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	repo, err := git.OpenRepo(absPath)
	if err != nil {
		return fmt.Errorf("not a git repository at %s.\nThe update command requires a git repository to detect changes.\nIf this is a new project, run 'goscribe generate' first", absPath)
	}

	stateFile := filepath.Join(absPath, ".goscribe-state")
	if _, statErr := os.Stat(stateFile); os.IsNotExist(statErr) {
		return fmt.Errorf("no previous documentation state found at %s.\nYou must run 'goscribe generate' before using 'update' to establish a baseline", stateFile)
	}

	lastCommit, err := docs.LoadCommitState(absPath)
	if err != nil {
		return fmt.Errorf("failed to load documentation state from %s.\nThe file may be corrupted. Try running 'goscribe generate' to reset it: %w", stateFile, err)
	}

	currentCommit, err := repo.GetCurrentCommit()
	if err != nil {
		return fmt.Errorf("get current commit: %w.\nCheck that the git repository is valid", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w.\nCheck your config file at ~/.goscribe.yaml", err)
	}

	var changedFiles []string
	if lastCommit != currentCommit {
		changedFiles, err = repo.GetChangedFiles(lastCommit, currentCommit)
		if err != nil {
			return fmt.Errorf("get changed files: %w", err)
		}
	}

	if len(changedFiles) == 0 {
		if cfg.CI {
			ci.Format(cmd.OutOrStdout(), ci.Result{
				Success:   true,
				Command:   "update",
				NoChanges: true,
			}, cfg.OutputFormat)
			return errExit(2)
		}
		fmt.Println("No changes detected since last documentation generation.")
		return nil
	}

	provider, err := ai.NewProvider(cfg)
	if err != nil {
		return err
	}

	updater := docs.NewUpdater(provider, cfg.OutputDir)

	if cfg.Profile != "" {
		profile, err := docs.ResolveProfile(cfg.Profile)
		if err != nil {
			return fmt.Errorf("resolve profile %q: %w", cfg.Profile, err)
		}
		updater.WithProfile(profile)
	}

	if cfg.Template != "" {
		tmpl, terr := docs.ResolveTemplate(cfg.Template)
		if terr != nil {
			return fmt.Errorf("resolve template %q: %w", cfg.Template, terr)
		}
		if tmpl != nil {
			updater.WithTemplate(tmpl)
		}
	}

	if err := updater.UpdateContext(cmd.Context(), absPath, changedFiles); err != nil {
		if cfg.CI {
			ci.Format(cmd.OutOrStdout(), ci.Result{
				Success: false,
				Command: "update",
				Error:   err.Error(),
			}, cfg.OutputFormat)
			return errExit(1)
		}
		return fmt.Errorf("update docs: %w", err)
	}

	if err := docs.SaveCommitState(absPath, currentCommit); err != nil {
		return fmt.Errorf("save commit state: %w", err)
	}

	if cfg.CI {
		ci.Format(cmd.OutOrStdout(), ci.Result{
			Success:      true,
			Command:      "update",
			FilesUpdated: len(changedFiles),
			FilesChanged: changedFiles,
			OutputDir:    cfg.OutputDir,
		}, cfg.OutputFormat)
		return nil
	}

	fmt.Printf("Documentation updated successfully (%d files changed)\n", len(changedFiles))
	return nil
}
