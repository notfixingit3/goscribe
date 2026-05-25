package opencode

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestNewFileWatcher(t *testing.T) {
	client := NewClient(DefaultConfig())
	fw := NewFileWatcher(client)

	if fw == nil {
		t.Fatal("NewFileWatcher returned nil")
	}
	if fw.client != client {
		t.Error("client mismatch")
	}
	if fw.debounceDelay != DebounceDelay {
		t.Errorf("expected default debounce delay %v, got %v", DebounceDelay, fw.debounceDelay)
	}
	if fw.pendingPaths == nil {
		t.Error("pendingPaths not initialized")
	}
}

func TestFileWatcher_SetDebounceDelay(t *testing.T) {
	fw := NewFileWatcher(NewClient(DefaultConfig()))

	fw.SetDebounceDelay(500 * time.Millisecond)
	if fw.debounceDelay != 500*time.Millisecond {
		t.Errorf("expected 500ms, got %v", fw.debounceDelay)
	}

	fw.SetDebounceDelay(0)
	if fw.debounceDelay != 500*time.Millisecond {
		t.Error("zero delay should be ignored")
	}

	fw.SetDebounceDelay(-1 * time.Second)
	if fw.debounceDelay != 500*time.Millisecond {
		t.Error("negative delay should be ignored")
	}
}

func TestFileWatcher_Watch_StartStop(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = true
	cfg.ProjectPath = tmpDir

	client := NewClient(cfg)
	ctx := context.Background()
	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer client.Shutdown(ctx)

	if client.watcher == nil {
		t.Fatal("expected watcher to be created")
	}

	client.watcher.Stop()
}

func TestFileWatcher_Watch_AlreadyRunning(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false
	client := NewClient(cfg)
	fw := NewFileWatcher(client)

	ctx := context.Background()
	if err := fw.Watch(ctx, tmpDir); err != nil {
		t.Fatalf("first Watch failed: %v", err)
	}
	defer fw.Stop()

	err := fw.Watch(ctx, tmpDir)
	if err == nil {
		t.Fatal("expected error when calling Watch twice")
	}
}

func TestFileWatcher_Stop_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false
	client := NewClient(cfg)
	fw := NewFileWatcher(client)

	ctx := context.Background()
	if err := fw.Watch(ctx, tmpDir); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	fw.Stop()
	fw.Stop()
}

func TestFileWatcher_Debounce(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false
	client := NewClient(cfg)

	var callCount int32
	client.RegisterHandler(EventFileSaved, func(_ context.Context, e Event) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	})

	fw := NewFileWatcher(client)
	fw.SetDebounceDelay(100 * time.Millisecond)

	ctx := context.Background()
	if err := fw.Watch(ctx, tmpDir); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}
	defer fw.Stop()

	filePath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(filePath, []byte("package main"), 0644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	fw.handleEvent(fsnotify.Event{Name: filePath, Op: fsnotify.Write})
	fw.handleEvent(fsnotify.Event{Name: filePath, Op: fsnotify.Write})
	fw.handleEvent(fsnotify.Event{Name: filePath, Op: fsnotify.Write})

	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 debounced event, got %d", callCount)
	}
}

func TestFileWatcher_Debounce_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false
	client := NewClient(cfg)

	var mu sync.Mutex
	paths := make(map[string]struct{})
	client.RegisterHandler(EventFileSaved, func(_ context.Context, e Event) error {
		mu.Lock()
		paths[e.FilePath] = struct{}{}
		mu.Unlock()
		return nil
	})

	fw := NewFileWatcher(client)
	fw.SetDebounceDelay(100 * time.Millisecond)

	ctx := context.Background()
	if err := fw.Watch(ctx, tmpDir); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}
	defer fw.Stop()

	file1 := filepath.Join(tmpDir, "a.go")
	file2 := filepath.Join(tmpDir, "b.go")
	os.WriteFile(file1, []byte("package main"), 0644)
	os.WriteFile(file2, []byte("package main"), 0644)

	fw.handleEvent(fsnotify.Event{Name: file1, Op: fsnotify.Write})
	fw.handleEvent(fsnotify.Event{Name: file2, Op: fsnotify.Write})

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	if len(paths) != 2 {
		t.Errorf("expected 2 distinct file events, got %d", len(paths))
	}
	mu.Unlock()
}

func TestFileWatcher_shouldWatch_Extensions(t *testing.T) {
	fw := NewFileWatcher(NewClient(DefaultConfig()))
	fw.projectPath = "/project"

	tests := []struct {
		path     string
		expected bool
	}{
		{"/project/main.go", true},
		{"/project/main_test.go", false},
		{"/project/readme.md", false},
		{"/project/main.js", false},
		{"/project/vendor/pkg.go", false},
		{"/project/node_modules/pkg.go", false},
		{"/project/.git/hooks.go", false},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.path), func(t *testing.T) {
			got := fw.shouldWatch(tt.path)
			if got != tt.expected {
				t.Errorf("shouldWatch(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestFileWatcher_shouldWatch_Gitignore(t *testing.T) {
	tmpDir := t.TempDir()

	gitignoreContent := `*.tmp
build/
/dist
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("write gitignore: %v", err)
	}

	fw := NewFileWatcher(NewClient(DefaultConfig()))
	fw.projectPath = tmpDir
	fw.gitignore = fw.loadGitignore(tmpDir)

	tests := []struct {
		path     string
		expected bool
	}{
		{filepath.Join(tmpDir, "main.go"), true},
		{filepath.Join(tmpDir, "main.tmp"), false},
		{filepath.Join(tmpDir, "build", "output.go"), true},
		{filepath.Join(tmpDir, "dist", "app.go"), true},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.path), func(t *testing.T) {
			got := fw.shouldWatch(tt.path)
			if got != tt.expected {
				t.Errorf("shouldWatch(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestFileWatcher_loadGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	content := `# comment
*.log

vendor/
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(content), 0644); err != nil {
		t.Fatalf("write gitignore: %v", err)
	}

	fw := NewFileWatcher(NewClient(DefaultConfig()))
	patterns := fw.loadGitignore(tmpDir)

	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
	if patterns[0] != "*.log" {
		t.Errorf("expected *.log, got %s", patterns[0])
	}
	if patterns[1] != "vendor/" {
		t.Errorf("expected vendor/, got %s", patterns[1])
	}
}

func TestFileWatcher_loadGitignore_Missing(t *testing.T) {
	tmpDir := t.TempDir()

	fw := NewFileWatcher(NewClient(DefaultConfig()))
	patterns := fw.loadGitignore(tmpDir)

	if patterns != nil {
		t.Errorf("expected nil for missing gitignore, got %v", patterns)
	}
}

func TestMatchGitignorePattern(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"foo.go", "*.go", true},
		{"foo.js", "*.go", false},
		{"dir/foo.go", "*.go", false},
		{"build/output.go", "build/", false},
		{"build/output.go", "build", true},
		{"src/build/output.go", "build/", false},
		{"dist/app.go", "/dist", false},
		{"src/dist/app.go", "/dist", false},
		{"any/deep/path.go", "**", true},
		{"src/vendor/pkg.go", "vendor/", false},
		{"foo_bar.go", "foo_*.go", true},
		{"test_foo.go", "*_foo.go", true},
		{"main.go", "main.go", true},
		{"main.go", "other.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.pattern, func(t *testing.T) {
			got := matchGitignorePattern(tt.path, tt.pattern)
			if got != tt.want {
				t.Errorf("matchGitignorePattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestFileWatcher_handleEvent_FilterNonWrite(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false
	client := NewClient(cfg)

	client.RegisterHandler(EventFileSaved, func(_ context.Context, _ Event) error {
		return nil
	})

	fw := NewFileWatcher(client)
	fw.SetDebounceDelay(50 * time.Millisecond)

	ctx := context.Background()
	if err := fw.Watch(ctx, tmpDir); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}
	defer fw.Stop()

	filePath := filepath.Join(tmpDir, "main.go")
	os.WriteFile(filePath, []byte("package main"), 0644)

	fw.handleEvent(fsnotify.Event{Name: filePath, Op: fsnotify.Chmod})

	time.Sleep(150 * time.Millisecond)

	fw.mu.Lock()
	pendingCount := len(fw.pendingPaths)
	fw.mu.Unlock()

	if pendingCount != 0 {
		t.Error("expected no pending paths for non-write event")
	}
}

func TestFileWatcher_addRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package main"), 0644)

	vendorDir := filepath.Join(tmpDir, "vendor")
	os.MkdirAll(vendorDir, 0755)
	os.WriteFile(filepath.Join(vendorDir, "pkg.go"), []byte("package pkg"), 0644)

	fw := NewFileWatcher(NewClient(DefaultConfig()))
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("create fsnotify watcher: %v", err)
	}
	fw.watcher = watcher
	fw.projectPath = tmpDir

	if err := fw.addRecursive(tmpDir); err != nil {
		t.Fatalf("addRecursive failed: %v", err)
	}

	watcher.Close()
}

func TestFileWatcher_shouldSkipDir(t *testing.T) {
	fw := NewFileWatcher(NewClient(DefaultConfig()))

	tests := []struct {
		rel  string
		base string
		want bool
	}{
		{".", ".", false},
		{"vendor", "vendor", true},
		{"node_modules", "node_modules", true},
		{".git", ".git", true},
		{".hidden", ".hidden", true},
		{"src", "src", false},
	}

	for _, tt := range tests {
		t.Run(tt.base, func(t *testing.T) {
			got := fw.shouldSkipDir(tt.rel, tt.base)
			if got != tt.want {
				t.Errorf("shouldSkipDir(%q, %q) = %v, want %v", tt.rel, tt.base, got, tt.want)
			}
		})
	}
}

func TestFileWatcher_flushPending(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false
	client := NewClient(cfg)

	var mu sync.Mutex
	var received []string
	client.RegisterHandler(EventFileSaved, func(_ context.Context, e Event) error {
		mu.Lock()
		received = append(received, e.FilePath)
		mu.Unlock()
		return nil
	})

	fw := NewFileWatcher(client)
	fw.projectPath = tmpDir
	fw.ctx = context.Background()
	fw.pendingPaths = map[string]struct{}{
		filepath.Join(tmpDir, "a.go"): {},
		filepath.Join(tmpDir, "b.go"): {},
	}

	fw.flushPending()

	mu.Lock()
	if len(received) != 2 {
		t.Errorf("expected 2 events, got %d", len(received))
	}
	mu.Unlock()

	fw.mu.Lock()
	if len(fw.pendingPaths) != 0 {
		t.Error("expected pendingPaths to be cleared")
	}
	if fw.timer != nil {
		t.Error("expected timer to be nil after flush")
	}
	fw.mu.Unlock()
}

func TestFileWatcher_isInDir(t *testing.T) {
	fw := NewFileWatcher(NewClient(DefaultConfig()))

	tests := []struct {
		path string
		dir  string
		want bool
	}{
		{"/project/vendor/pkg.go", "vendor", true},
		{"/project/src/main.go", "vendor", false},
		{"/project/node_modules/lib.js", "node_modules", true},
		{"/project/.git/hooks", ".git", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := fw.isInDir(tt.path, tt.dir)
			if got != tt.want {
				t.Errorf("isInDir(%q, %q) = %v, want %v", tt.path, tt.dir, got, tt.want)
			}
		})
	}
}

func TestFileWatcher_Debounce_ResetTimer(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false
	client := NewClient(cfg)

	var callCount int32
	client.RegisterHandler(EventFileSaved, func(_ context.Context, e Event) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	})

	fw := NewFileWatcher(client)
	fw.SetDebounceDelay(200 * time.Millisecond)

	ctx := context.Background()
	if err := fw.Watch(ctx, tmpDir); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}
	defer fw.Stop()

	filePath := filepath.Join(tmpDir, "main.go")
	os.WriteFile(filePath, []byte("package main"), 0644)

	fw.handleEvent(fsnotify.Event{Name: filePath, Op: fsnotify.Write})
	time.Sleep(100 * time.Millisecond)
	fw.handleEvent(fsnotify.Event{Name: filePath, Op: fsnotify.Write})
	time.Sleep(100 * time.Millisecond)
	fw.handleEvent(fsnotify.Event{Name: filePath, Op: fsnotify.Write})

	time.Sleep(400 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 debounced event after resets, got %d", callCount)
	}
}

func TestFileWatcher_Integration_RealFileChange(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false
	client := NewClient(cfg)

	var mu sync.Mutex
	var events []Event
	client.RegisterHandler(EventFileSaved, func(_ context.Context, e Event) error {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
		return nil
	})

	fw := NewFileWatcher(client)
	fw.SetDebounceDelay(100 * time.Millisecond)

	ctx := context.Background()
	if err := fw.Watch(ctx, tmpDir); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}
	defer fw.Stop()

	filePath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(filePath, []byte("package main"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := os.WriteFile(filePath, []byte("package main\n\nfunc main() {}"), 0644); err != nil {
		t.Fatalf("update file: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	if len(events) < 1 {
		t.Errorf("expected at least 1 event from real file change, got %d", len(events))
	}
	mu.Unlock()
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		segment string
		pattern string
		want    bool
	}{
		{"foo", "*", true},
		{"foo", "foo", true},
		{"foo", "bar", false},
		{"foobar", "foo*", true},
		{"foobar", "*bar", true},
		{"foobar", "foo*baz", false},
		{"main.go", "*.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.segment+"_"+tt.pattern, func(t *testing.T) {
			got := matchGlob(tt.segment, tt.pattern)
			if got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.segment, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMatchSegments(t *testing.T) {
	tests := []struct {
		path         string
		parts        []string
		anchored     bool
		isDirPattern bool
		want         bool
	}{
		{"foo/bar", []string{"foo", "bar"}, false, false, true},
		{"foo/bar/baz", []string{"foo", "bar"}, false, false, true},
		{"foo/bar", []string{"foo", "bar", "baz"}, false, false, false},
		{"foo/bar", []string{"foo"}, false, false, true},
		{"foo/bar", []string{"bar"}, false, false, false},
		{"foo/bar", []string{"bar"}, true, false, false},
		{"foo/bar", []string{"foo", "bar"}, true, false, true},
		{"foo/bar/", []string{"foo", "bar"}, false, true, true},
		{"foo/bar", []string{"foo", "bar"}, false, true, true},
		{"foo", []string{"**"}, false, false, true},
		{"foo/bar", []string{"**", "bar"}, false, false, true},
		{"foo/bar/baz", []string{"**", "baz"}, false, false, true},
		{"foo/bar", []string{"**", "qux"}, false, false, false},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.parts, "/"), func(t *testing.T) {
			got := matchSegments(tt.path, tt.parts, tt.anchored, tt.isDirPattern)
			if got != tt.want {
				t.Errorf("matchSegments(%q, %v, %v, %v) = %v, want %v",
					tt.path, tt.parts, tt.anchored, tt.isDirPattern, got, tt.want)
			}
		})
	}
}
