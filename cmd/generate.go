// Package cmd implements the goscribe CLI commands.
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
	_ = viper.BindPFlag("force", generateCmd.Flags().Lookup("force"))

	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	sourcePath := "."
	if len(args) > 0 {
		sourcePath = args[0]
	}

	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("resolve path %q: %w.\nCheck that the path is valid", sourcePath, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s\nCheck the path and try again", absPath)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied accessing %s\nCheck file permissions", absPath)
		}
		return fmt.Errorf("cannot access path %s: %w", absPath, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s\nProvide a directory containing source code, not a file", absPath)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w.\nCheck your config file at ~/.goscribe.yaml", err)
	}

	if cfg.OutputDir == "" {
		return fmt.Errorf("output directory is empty in config.\nSet it with: goscribe generate --output docs")
	}

	provider, err := ai.NewProvider(cfg)
	if err != nil {
		return err
	}

	generator := docs.NewGenerator(provider, cfg.OutputDir)

	if cfg.Profile != "" {
		profile, perr := docs.ResolveProfile(cfg.Profile)
		if perr != nil {
			return fmt.Errorf("resolve profile %q: %w", cfg.Profile, perr)
		}
		generator.WithProfile(profile)
	}

	if cfg.Template != "" {
		tmpl, terr := docs.ResolveTemplate(cfg.Template)
		if terr != nil {
			return fmt.Errorf("resolve template %q: %w", cfg.Template, terr)
		}
		if tmpl != nil {
			generator.WithTemplate(tmpl)
		}
	}

	if genErr := generator.GenerateContext(cmd.Context(), absPath); genErr != nil {
		if cfg.CI {
			ci.Format(cmd.OutOrStdout(), ci.Result{
				Success: false,
				Command: "generate",
				Error:   genErr.Error(),
			}, cfg.OutputFormat)
			return errExit(1)
		}
		return fmt.Errorf("generate docs: %w", genErr)
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

	if cfg.CI {
		ci.Format(cmd.OutOrStdout(), ci.Result{
			Success:        true,
			Command:        "generate",
			FilesGenerated: generator.FileCount(),
			OutputDir:      cfg.OutputDir,
		}, cfg.OutputFormat)
		return nil
	}

	fmt.Printf("Documentation generated successfully in %s\n", cfg.OutputDir)
	return nil
}
