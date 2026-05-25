package goscribe

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/house/goscribe/internal/docs"
)

// GenerateOptions controls the behavior of the Generate method.
type GenerateOptions struct {
	OutputDir     string
	Force         bool
	SkipStateSave bool
}

// GenerateResult holds the outcome of a documentation generation run.
type GenerateResult struct {
	SourcePath     string
	OutputDir      string
	FilesGenerated int
	StateSaved     bool
	Commit         string
}

// Generate produces documentation for all source files in sourcePath.
// It respects context cancellation and saves git state unless SkipStateSave is set.
func (c *Client) Generate(ctx context.Context, sourcePath string, opts GenerateOptions) (*GenerateResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, &Error{
			Op:   "generate",
			Kind: ErrInvalidContext,
			Err:  err,
		}
	}

	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, &Error{
			Op:   "generate",
			Kind: ErrInvalidPath,
			Path: sourcePath,
			Err:  err,
		}
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &Error{
				Op:   "generate",
				Kind: ErrInvalidPath,
				Path: absPath,
				Err:  fmt.Errorf("path does not exist"),
			}
		}
		if os.IsPermission(err) {
			return nil, &Error{
				Op:   "generate",
				Kind: ErrInvalidPath,
				Path: absPath,
				Err:  fmt.Errorf("permission denied"),
			}
		}
		return nil, &Error{
			Op:   "generate",
			Kind: ErrInvalidPath,
			Path: absPath,
			Err:  err,
		}
	}

	if !info.IsDir() {
		return nil, &Error{
			Op:   "generate",
			Kind: ErrInvalidPath,
			Path: absPath,
			Err:  fmt.Errorf("path is not a directory"),
		}
	}

	outputDir := c.resolveOutputDir(opts.OutputDir)

	generator := docs.NewGenerator(c.provider, outputDir).
		WithWorkers(c.workers)
	if c.cacheDir != "" {
		generator = generator.WithCache(docs.NewCache(c.cacheDir))
	}
	if c.profile != "" {
		profile, err := docs.ResolveProfile(c.profile)
		if err != nil {
			return nil, &Error{
				Op:   "generate",
				Kind: ErrGenerationFailed,
				Path: absPath,
				Err:  fmt.Errorf("resolve profile: %w", err),
			}
		}
		generator.WithProfile(profile)
	}
	if c.template != "" {
		tmpl, err := docs.ResolveTemplate(c.template)
		if err != nil {
			return nil, &Error{
				Op:   "generate",
				Kind: ErrGenerationFailed,
				Path: absPath,
				Err:  fmt.Errorf("resolve template: %w", err),
			}
		}
		if tmpl != nil {
			generator.WithTemplate(tmpl)
		}
	}
	if err := generator.GenerateContext(ctx, absPath); err != nil {
		return nil, &Error{
			Op:   "generate",
			Kind: ErrGenerationFailed,
			Path: absPath,
			Err:  err,
		}
	}

	result := &GenerateResult{
		SourcePath:     absPath,
		OutputDir:      outputDir,
		FilesGenerated: generator.FileCount(),
	}

	if !opts.SkipStateSave {
		commit, saved, err := c.saveStateIfGit(absPath)
		if err != nil {
			return nil, &Error{
				Op:   "generate",
				Kind: ErrGenerationFailed,
				Path: absPath,
				Err:  fmt.Errorf("save state: %w", err),
			}
		}
		result.StateSaved = saved
		result.Commit = commit
	}

	return result, nil
}
