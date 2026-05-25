package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewMockRepo_CreatesValidRepo(t *testing.T) {
	repo, err := NewMockRepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer repo.Cleanup()

	if repo.Path == "" {
		t.Fatal("expected non-empty path")
	}

	gitDir := filepath.Join(repo.Path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatal(".git directory not found")
	}

	head, err := repo.HeadCommit()
	if err != nil {
		t.Fatalf("unexpected error getting head: %v", err)
	}
	if head == "" {
		t.Fatal("expected non-empty commit hash")
	}
}

func TestMockRepo_AddFile(t *testing.T) {
	repo, err := NewMockRepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer repo.Cleanup()

	firstCommit, _ := repo.HeadCommit()

	err = repo.AddFile("foo.go", "package foo\n")
	if err != nil {
		t.Fatalf("unexpected error adding file: %v", err)
	}

	secondCommit, _ := repo.HeadCommit()
	if secondCommit == firstCommit {
		t.Fatal("expected commit hash to change after AddFile")
	}

	content, err := os.ReadFile(filepath.Join(repo.Path, "foo.go"))
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}
	if string(content) != "package foo\n" {
		t.Errorf("expected %q, got %q", "package foo\n", string(content))
	}
}

func TestMockRepo_Cleanup(t *testing.T) {
	repo, err := NewMockRepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := repo.Path
	repo.Cleanup()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expected directory to be removed after cleanup")
	}
}
