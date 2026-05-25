// Package docs handles documentation generation and incremental updates.
package docs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/house/goscribe/internal/ai"
	"golang.org/x/sync/errgroup"
)

// Generator produces documentation from source files using an AI provider.
type Generator struct {
	provider  ai.Provider
	outputDir string
	fileCount int
	workers   int
	cache     *Cache
	profile   Profile
	template  Template
}

// NewGenerator creates a new documentation Generator.
func NewGenerator(provider ai.Provider, outputDir string) *Generator {
	return &Generator{
		provider:  provider,
		outputDir: outputDir,
		workers:   1,
		profile:   DefaultProfile(),
	}
}

// WithWorkers sets the number of concurrent workers for generation.
func (g *Generator) WithWorkers(n int) *Generator {
	if n < 1 {
		n = 1
	}
	g.workers = n
	return g
}

// WithCache attaches a Cache to the Generator for content-addressed caching.
func (g *Generator) WithCache(c *Cache) *Generator {
	g.cache = c
	return g
}

// WithProfile sets the documentation profile for generation.
func (g *Generator) WithProfile(profile Profile) *Generator {
	if profile != nil {
		g.profile = profile
	}
	return g
}

// WithTemplate sets the documentation template for generation.
func (g *Generator) WithTemplate(t Template) *Generator {
	g.template = t
	return g
}

func (g *Generator) templateName() string {
	if g.template == nil {
		return ""
	}
	return g.template.Name()
}

// FileCount returns the number of files processed in the last Generate call.
func (g *Generator) FileCount() int {
	return g.fileCount
}

// Generate walks the source path, sends each file to the AI provider, and writes docs.
func (g *Generator) Generate(sourcePath string) error {
	return g.GenerateContext(context.Background(), sourcePath)
}

// GenerateContext is like Generate but respects the given context for cancellation.
func (g *Generator) GenerateContext(ctx context.Context, sourcePath string) error {
	if err := os.MkdirAll(g.outputDir, 0750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	files, err := g.collectSourceFiles(sourcePath)
	if err != nil {
		return fmt.Errorf("collect files: %w", err)
	}

	g.fileCount = 0

	if g.workers <= 1 {
		for _, file := range files {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("generate canceled: %w", err)
			}
			if err := g.processFileContext(ctx, sourcePath, file); err != nil {
				return fmt.Errorf("process %s: %w", file, err)
			}
			g.fileCount++
		}
		return nil
	}

	var count atomic.Int32
	grp, ctx := errgroup.WithContext(ctx)
	grp.SetLimit(g.workers)

	for _, file := range files {
		file := file
		grp.Go(func() error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("generate canceled: %w", err)
			}
			if err := g.processFileContext(ctx, sourcePath, file); err != nil {
				return fmt.Errorf("process %s: %w", file, err)
			}
			count.Add(1)
			return nil
		})
	}

	if err := grp.Wait(); err != nil {
		return err
	}

	g.fileCount = int(count.Load())
	return nil
}

func (g *Generator) collectSourceFiles(sourcePath string) ([]string, error) {
	var files []string

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".md") {
			relPath, err := filepath.Rel(sourcePath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}

func (g *Generator) processFileContext(ctx context.Context, sourcePath, file string) error {
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

	content, err := os.ReadFile(absJoined) // #nosec G304 -- path validated against traversal above
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if g.cache != nil {
		if cached, ok := g.cache.Get(g.profile.Name(), g.templateName(), content); ok {
			doc := cached
			outputFile := filepath.Join(g.outputDir, cleanFile+".md")
			if mkErr := os.MkdirAll(filepath.Dir(outputFile), 0750); mkErr != nil {
				return fmt.Errorf("create dir: %w", mkErr)
			}
			if wrErr := os.WriteFile(outputFile, []byte(doc), 0600); wrErr != nil {
				return fmt.Errorf("write doc: %w", wrErr)
			}
			return nil
		}
	}

	b := NewPromptBuilder(g.profile).WithTemplate(g.template)
	prompt := b.BuildGeneratePrompt(file, content)

	doc, err := g.provider.Generate(ctx, prompt)
	if err != nil {
		return fmt.Errorf("generate doc: %w", err)
	}

	doc = b.FormatOutput(doc)

	if g.cache != nil {
		_ = g.cache.Set(g.profile.Name(), g.templateName(), content, doc)
	}

	outputFile := filepath.Join(g.outputDir, cleanFile+".md")
	if err := os.MkdirAll(filepath.Dir(outputFile), 0750); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	if err := os.WriteFile(outputFile, []byte(doc), 0600); err != nil {
		return fmt.Errorf("write doc: %w", err)
	}

	return nil
}
