package docs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	content := []byte("package main\n\nfunc main() {}\n")
	doc := "# main.go\n\nThis is the main package."

	// Set should succeed
	if err := cache.Set(DefaultProfileName, content, doc); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get should return the cached doc
	got, ok := cache.Get(DefaultProfileName, content)
	if !ok {
		t.Fatal("Get returned false for existing key")
	}
	if got != doc {
		t.Errorf("Get returned %q, want %q", got, doc)
	}
}

func TestCacheMiss(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	content := []byte("package main\n")

	got, ok := cache.Get(DefaultProfileName, content)
	if ok {
		t.Errorf("Get returned true for missing key, got %q", got)
	}
}

func TestCacheDifferentContent(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	content1 := []byte("package main\n")
	content2 := []byte("package foo\n")

	if err := cache.Set(DefaultProfileName, content1, "doc for main"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// content2 should be a cache miss
	got, ok := cache.Get(DefaultProfileName, content2)
	if ok {
		t.Errorf("Get returned true for different content, got %q", got)
	}

	// content1 should still be a hit
	got, ok = cache.Get(DefaultProfileName, content1)
	if !ok {
		t.Fatal("Get returned false for existing key")
	}
	if got != "doc for main" {
		t.Errorf("Get returned %q, want %q", got, "doc for main")
	}
}

func TestCacheKeyDeterminism(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	content := []byte("package main\n")

	if err := cache.Set(DefaultProfileName, content, "doc"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Same content should produce same key
	got, ok := cache.Get(DefaultProfileName, []byte("package main\n"))
	if !ok {
		t.Fatal("Get returned false for same content")
	}
	if got != "doc" {
		t.Errorf("Get returned %q, want %q", got, "doc")
	}
}

func TestCachePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	cache1 := NewCache(tmpDir)

	content := []byte("package main\n")
	if err := cache1.Set(DefaultProfileName, content, "persistent doc"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// New cache instance pointing to same dir
	cache2 := NewCache(tmpDir)
	got, ok := cache2.Get(DefaultProfileName, content)
	if !ok {
		t.Fatal("Get returned false for persisted entry")
	}
	if got != "persistent doc" {
		t.Errorf("Get returned %q, want %q", got, "persistent doc")
	}
}

func TestCacheConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	// Write multiple entries concurrently
	done := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			content := []byte(string(rune('a' + i)))
			_ = cache.Set(DefaultProfileName, content, string(rune('A'+i)))
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all entries
	for i := 0; i < 10; i++ {
		content := []byte(string(rune('a' + i)))
		got, ok := cache.Get(DefaultProfileName, content)
		if !ok {
			t.Errorf("Get returned false for entry %d", i)
			continue
		}
		want := string(rune('A' + i))
		if got != want {
			t.Errorf("Entry %d: got %q, want %q", i, got, want)
		}
	}
}

func TestCacheSetCreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "nested", "cache")
	cache := NewCache(cacheDir)

	content := []byte("package main\n")
	if err := cache.Set(DefaultProfileName, content, "doc"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify dir was created
	if _, err := os.Stat(cacheDir); err != nil {
		t.Errorf("Cache dir not created: %v", err)
	}
}
