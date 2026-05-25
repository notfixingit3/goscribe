package docs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/house/goscribe/internal/ai"
	"golang.org/x/sync/errgroup"
)

// Updater incrementally updates documentation for changed files.
type Updater struct {
	provider  ai.Provider
	outputDir string
	workers   int
	cache     *Cache
	profile   Profile
	template  Template
}

// NewUpdater creates a new documentation Updater.
func NewUpdater(provider ai.Provider, outputDir string) *Updater {
	return &Updater{
		provider:  provider,
		outputDir: outputDir,
		workers:   1,
		profile:   DefaultProfile(),
	}
}

// WithWorkers sets the number of concurrent workers for updates.
func (u *Updater) WithWorkers(n int) *Updater {
	if n < 1 {
		n = 1
	}
	u.workers = n
	return u
}

// WithCache attaches a Cache to the Updater for content-addressed caching.
func (u *Updater) WithCache(c *Cache) *Updater {
	u.cache = c
	return u
}

// WithProfile sets the documentation profile for updates.
func (u *Updater) WithProfile(profile Profile) *Updater {
	if profile != nil {
		u.profile = profile
	}
	return u
}

// WithTemplate sets the documentation template for updates.
func (u *Updater) WithTemplate(t Template) *Updater {
	u.template = t
	return u
}

func (u *Updater) templateName() string {
	if u.template == nil {
		return ""
	}
	return u.template.Name()
}

// Update regenerates documentation for each changed file.
func (u *Updater) Update(sourcePath string, changedFiles []string) error {
	return u.UpdateContext(context.Background(), sourcePath, changedFiles)
}

// UpdateContext is like Update but respects the given context for cancellation.
func (u *Updater) UpdateContext(ctx context.Context, sourcePath string, changedFiles []string) error {
	if u.workers <= 1 {
		for _, file := range changedFiles {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("update canceled: %w", err)
			}
			if err := u.updateFileContext(ctx, sourcePath, file); err != nil {
				return fmt.Errorf("update %s: %w", file, err)
			}
		}
		return nil
	}

	grp, ctx := errgroup.WithContext(ctx)
	grp.SetLimit(u.workers)

	for _, file := range changedFiles {
		file := file
		grp.Go(func() error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("update canceled: %w", err)
			}
			if err := u.updateFileContext(ctx, sourcePath, file); err != nil {
				return fmt.Errorf("update %s: %w", file, err)
			}
			return nil
		})
	}

	return grp.Wait()
}

func (u *Updater) updateFileContext(ctx context.Context, sourcePath, file string) error {
	cleanFile := filepath.Clean(file)
	joined := filepath.Join(sourcePath, cleanFile)
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return fmt.Errorf("resolve file path: %w", err)
	}
	if !strings.HasPrefix(absJoined, absSource+string(os.PathSeparator)) && absJoined != absSource {
		return fmt.Errorf("path traversal detected: %s escapes source directory", file)
	}

	content, err := os.ReadFile(absJoined)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if u.cache != nil {
		if cached, ok := u.cache.Get(u.profile.Name(), u.templateName(), content); ok {
			doc := cached
			outputFile := filepath.Join(u.outputDir, cleanFile+".md")
			if mkErr := os.MkdirAll(filepath.Dir(outputFile), 0750); mkErr != nil {
				return fmt.Errorf("create dir: %w", mkErr)
			}
			if wrErr := os.WriteFile(outputFile, []byte(doc), 0600); wrErr != nil {
				return fmt.Errorf("write doc: %w", wrErr)
			}
			return nil
		}
	}

	b := NewPromptBuilder(u.profile).WithTemplate(u.template)
	prompt := b.BuildUpdatePrompt(file, content)

	doc, err := u.provider.Generate(ctx, prompt)
	if err != nil {
		return fmt.Errorf("generate doc: %w", err)
	}

	doc = b.FormatOutput(doc)

	if u.cache != nil {
		_ = u.cache.Set(u.profile.Name(), u.templateName(), content, doc)
	}

	outputFile := filepath.Join(u.outputDir, cleanFile+".md")
	if err := os.MkdirAll(filepath.Dir(outputFile), 0750); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	if err := os.WriteFile(outputFile, []byte(doc), 0600); err != nil {
		return fmt.Errorf("write doc: %w", err)
	}

	return nil
}
