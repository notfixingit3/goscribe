package goscribe

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/house/goscribe/internal/docs"
	"github.com/house/goscribe/internal/git"
)

// UpdateOptions controls the behavior of the Update method.
type UpdateOptions struct {
	OutputDir     string
	ChangedFiles  []string
	FromCommit    string
	ToCommit      string
	SkipStateSave bool
}

// UpdateResult holds the outcome of a documentation update run.
type UpdateResult struct {
	SourcePath   string
	OutputDir    string
	FromCommit   string
	ToCommit     string
	ChangedFiles []string
	FilesUpdated int
	NoChanges    bool
	StateSaved   bool
}

// Update incrementally updates documentation for files changed since the last generation.
// It respects context cancellation and saves git state unless SkipStateSave is set.
func (c *Client) Update(ctx context.Context, sourcePath string, opts UpdateOptions) (*UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, &Error{
			Op:   "update",
			Kind: ErrInvalidContext,
			Err:  err,
		}
	}

	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, &Error{
			Op:   "update",
			Kind: ErrInvalidPath,
			Path: sourcePath,
			Err:  err,
		}
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &Error{
				Op:   "update",
				Kind: ErrInvalidPath,
				Path: absPath,
				Err:  fmt.Errorf("path does not exist"),
			}
		}
		return nil, &Error{
			Op:   "update",
			Kind: ErrInvalidPath,
			Path: absPath,
			Err:  err,
		}
	}

	if !info.IsDir() {
		return nil, &Error{
			Op:   "update",
			Kind: ErrInvalidPath,
			Path: absPath,
			Err:  fmt.Errorf("path is not a directory"),
		}
	}

	repo, err := git.OpenRepo(absPath)
	if err != nil {
		return nil, &Error{
			Op:   "update",
			Kind: ErrNotGitRepository,
			Path: absPath,
			Err:  err,
		}
	}

	stateFile := filepath.Join(absPath, ".goscribe-state")
	if _, statErr := os.Stat(stateFile); os.IsNotExist(statErr) {
		return nil, &Error{
			Op:   "update",
			Kind: ErrStateNotFound,
			Path: stateFile,
			Err:  fmt.Errorf("state file not found"),
		}
	}

	fromCommit := opts.FromCommit
	if fromCommit == "" {
		fromCommit, err = docs.LoadCommitState(absPath)
		if err != nil {
			return nil, &Error{
				Op:   "update",
				Kind: ErrStateInvalid,
				Path: stateFile,
				Err:  err,
			}
		}
	}

	toCommit := opts.ToCommit
	if toCommit == "" {
		toCommit, err = repo.GetCurrentCommit()
		if err != nil {
			return nil, &Error{
				Op:   "update",
				Kind: ErrUpdateFailed,
				Path: absPath,
				Err:  err,
			}
		}
	}

	if fromCommit == toCommit {
		return &UpdateResult{
			SourcePath: absPath,
			OutputDir:  c.resolveOutputDir(opts.OutputDir),
			FromCommit: fromCommit,
			ToCommit:   toCommit,
			NoChanges:  true,
		}, nil
	}

	changedFiles := opts.ChangedFiles
	if len(changedFiles) == 0 {
		changedFiles, err = repo.GetChangedFiles(fromCommit, toCommit)
		if err != nil {
			return nil, &Error{
				Op:   "update",
				Kind: ErrUpdateFailed,
				Path: absPath,
				Err:  err,
			}
		}
	}

	if len(changedFiles) == 0 {
		return &UpdateResult{
			SourcePath: absPath,
			OutputDir:  c.resolveOutputDir(opts.OutputDir),
			FromCommit: fromCommit,
			ToCommit:   toCommit,
			NoChanges:  true,
		}, nil
	}

	outputDir := c.resolveOutputDir(opts.OutputDir)
	updater := docs.NewUpdater(c.provider, outputDir).
		WithWorkers(c.workers)
	if c.cacheDir != "" {
		updater = updater.WithCache(docs.NewCache(c.cacheDir))
	}
	if c.profile != "" {
		profile, err := docs.ResolveProfile(c.profile)
		if err != nil {
			return nil, &Error{
				Op:   "update",
				Kind: ErrUpdateFailed,
				Path: absPath,
				Err:  fmt.Errorf("resolve profile: %w", err),
			}
		}
		updater.WithProfile(profile)
	}
	if err := updater.UpdateContext(ctx, absPath, changedFiles); err != nil {
		return nil, &Error{
			Op:   "update",
			Kind: ErrUpdateFailed,
			Path: absPath,
			Err:  err,
		}
	}

	result := &UpdateResult{
		SourcePath:   absPath,
		OutputDir:    outputDir,
		FromCommit:   fromCommit,
		ToCommit:     toCommit,
		ChangedFiles: changedFiles,
		FilesUpdated: len(changedFiles),
	}

	if !opts.SkipStateSave {
		if err := docs.SaveCommitState(absPath, toCommit); err != nil {
			return nil, &Error{
				Op:   "update",
				Kind: ErrUpdateFailed,
				Path: absPath,
				Err:  fmt.Errorf("save state: %w", err),
			}
		}
		result.StateSaved = true
	}

	return result, nil
}
