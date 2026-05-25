// Package opencode provides the OpenCode plugin client for integrating GoScribe
// documentation generation into IDE and editor environments.
package opencode

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// DebounceDelay is the default duration to wait for file changes to settle
// before triggering documentation generation.
const DebounceDelay = 2 * time.Second

// FileWatcher watches a project directory for Go file changes, debounces
// rapid changes, and dispatches EventFileSaved events through the plugin
// client. It respects .gitignore patterns and excludes test files, vendor,
// and node_modules directories.
type FileWatcher struct {
	client        *Client
	debounceDelay time.Duration
	watcher       *fsnotify.Watcher
	gitignore     []string
	projectPath   string
	mu            sync.Mutex
	timer         *time.Timer
	pendingPaths  map[string]struct{}
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewFileWatcher creates a new FileWatcher associated with the given plugin
// client. The debounce delay defaults to DebounceDelay and can be overridden
// by setting a positive delay on the returned watcher.
func NewFileWatcher(client *Client) *FileWatcher {
	return &FileWatcher{
		client:        client,
		debounceDelay: DebounceDelay,
		pendingPaths:  make(map[string]struct{}),
	}
}

// SetDebounceDelay configures the debounce duration. Values less than or
// equal to zero are ignored.
func (fw *FileWatcher) SetDebounceDelay(d time.Duration) {
	if d <= 0 {
		return
	}
	fw.debounceDelay = d
}

// Watch starts watching the project directory for Go file changes. It loads
// .gitignore patterns if present and recursively adds directories to the
// underlying fsnotify watcher. This method blocks until the watcher is fully
// initialized or an error occurs; after that, events are processed in a
// background goroutine.
func (fw *FileWatcher) Watch(ctx context.Context, projectPath string) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.watcher != nil {
		return fmt.Errorf("file watcher already running")
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}
	fw.projectPath = absPath

	fw.gitignore = fw.loadGitignore(absPath)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create fsnotify watcher: %w", err)
	}
	fw.watcher = watcher

	fw.ctx, fw.cancel = context.WithCancel(ctx)

	if err := fw.addRecursive(absPath); err != nil {
		_ = watcher.Close()
		fw.watcher = nil
		return fmt.Errorf("add watch paths: %w", err)
	}

	fw.wg.Add(1)
	go fw.run()

	return nil
}

// Stop signals the watcher to shut down and waits for the background goroutine
// to finish. It is safe to call multiple times.
func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	if fw.cancel != nil {
		fw.cancel()
	}
	if fw.watcher != nil {
		_ = fw.watcher.Close()
	}
	fw.mu.Unlock()

	fw.wg.Wait()
}

// run is the background goroutine that processes fsnotify events and manages
// the debounce timer.
func (fw *FileWatcher) run() {
	defer fw.wg.Done()

	for {
		select {
		case <-fw.ctx.Done():
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			// Log or surface the error through the client if desired.
			_ = err
		}
	}
}

// handleEvent filters fsnotify events and schedules debounced dispatch.
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
		return
	}

	path := event.Name
	if !fw.shouldWatch(path) {
		return
	}

	fw.mu.Lock()
	fw.pendingPaths[path] = struct{}{}

	if fw.timer != nil {
		fw.timer.Stop()
	}
	fw.timer = time.AfterFunc(fw.debounceDelay, func() {
		fw.flushPending()
	})
	fw.mu.Unlock()
}

// flushPending dispatches EventFileSaved events for all pending paths and
// clears the pending set. It is called by the debounce timer.
func (fw *FileWatcher) flushPending() {
	fw.mu.Lock()
	paths := make([]string, 0, len(fw.pendingPaths))
	for p := range fw.pendingPaths {
		paths = append(paths, p)
	}
	fw.pendingPaths = make(map[string]struct{})
	fw.timer = nil
	fw.mu.Unlock()

	for _, path := range paths {
		event := Event{
			Type:        EventFileSaved,
			FilePath:    path,
			ProjectPath: fw.projectPath,
			Timestamp:   time.Now(),
		}
		fw.client.dispatchEvent(fw.ctx, event)
	}
}

// shouldWatch returns true if the given path should be monitored based on
// extension, test file naming, and ignore patterns.
func (fw *FileWatcher) shouldWatch(path string) bool {
	base := filepath.Base(path)

	// Only watch .go files.
	if filepath.Ext(path) != ".go" {
		return false
	}

	// Exclude test files.
	if strings.HasSuffix(base, "_test.go") {
		return false
	}

	// Exclude vendor and node_modules directories.
	if fw.isInDir(path, "vendor") || fw.isInDir(path, "node_modules") {
		return false
	}

	// Exclude .git directory.
	if fw.isInDir(path, ".git") {
		return false
	}

	// Exclude paths matched by .gitignore patterns.
	rel, err := filepath.Rel(fw.projectPath, path)
	if err != nil {
		return false
	}
	for _, pattern := range fw.gitignore {
		if matchGitignorePattern(rel, pattern) {
			return false
		}
	}

	return true
}

// isInDir reports whether path resides inside a directory named dirName.
func (fw *FileWatcher) isInDir(path, dirName string) bool {
	for part := range strings.SplitSeq(filepath.ToSlash(path), "/") {
		if part == dirName {
			return true
		}
	}
	return false
}

// addRecursive walks the project tree and adds all directories to the
// fsnotify watcher, skipping ignored paths.
func (fw *FileWatcher) addRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths we can't access.
		}
		if !info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		// Skip hidden directories, vendor, node_modules, and .git.
		if fw.shouldSkipDir(rel, filepath.Base(path)) {
			return filepath.SkipDir
		}

		// Skip directories ignored by .gitignore.
		for _, pattern := range fw.gitignore {
			if matchGitignorePattern(rel+"/", pattern) || matchGitignorePattern(rel, pattern) {
				return filepath.SkipDir
			}
		}

		if err := fw.watcher.Add(path); err != nil {
			return fmt.Errorf("watch %s: %w", path, err)
		}
		return nil
	})
}

// shouldSkipDir returns true for directories that should never be watched.
func (fw *FileWatcher) shouldSkipDir(rel, base string) bool {
	if base == "vendor" || base == "node_modules" || base == ".git" {
		return true
	}
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}
	if rel == "." {
		return false
	}
	return false
}

// loadGitignore reads .gitignore from the project root and returns the
// non-empty, non-comment patterns.
func (fw *FileWatcher) loadGitignore(projectPath string) []string {
	path := filepath.Join(projectPath, ".gitignore")
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, filepath.Clean(projectPath)) {
		return nil
	}
	f, err := os.Open(path) // #nosec G304 -- path is cleaned and validated against project root
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

// matchGitignorePattern performs a simple glob-style match for gitignore
// patterns. It supports directory wildcards (**) and single-segment wildcards
// (*). Returns true if the path matches the pattern.
func matchGitignorePattern(path, pattern string) bool {
	// Normalize separators.
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	// Directory-only patterns end with "/".
	isDirPattern := strings.HasSuffix(pattern, "/")
	if isDirPattern {
		pattern = strings.TrimSuffix(pattern, "/")
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
	}

	// Handle leading "/" (anchored to root).
	anchored := strings.HasPrefix(pattern, "/")
	if anchored {
		pattern = strings.TrimPrefix(pattern, "/")
	}

	parts := strings.Split(pattern, "/")
	return matchSegments(path, parts, anchored, isDirPattern)
}

// matchSegments recursively matches path segments against pattern parts.
func matchSegments(path string, parts []string, anchored, isDirPattern bool) bool {
	if len(parts) == 0 {
		return path == "" || path == "/" || (!anchored && !isDirPattern)
	}

	if parts[0] == "**" {
		// "**" matches zero or more directories.
		if len(parts) == 1 {
			return true
		}
		segments := strings.Split(path, "/")
		for i := 0; i <= len(segments); i++ {
			subPath := strings.Join(segments[i:], "/")
			if matchSegments(subPath, parts[1:], anchored, isDirPattern) {
				return true
			}
		}
		return false
	}

	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		return false
	}

	if !matchGlob(segments[0], parts[0]) {
		return false
	}

	remainingPath := strings.Join(segments[1:], "/")
	return matchSegments(remainingPath, parts[1:], anchored, isDirPattern)
}

// matchGlob returns true if the segment matches the glob pattern.
func matchGlob(segment, pattern string) bool {
	// Simple glob: only "*" wildcard supported.
	if pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return segment == pattern
	}

	parts := strings.Split(pattern, "*")
	if len(parts) == 2 {
		return strings.HasPrefix(segment, parts[0]) && strings.HasSuffix(segment, parts[1])
	}

	// Fallback for complex patterns: use filepath.Match.
	matched, _ := filepath.Match(pattern, segment)
	return matched
}
