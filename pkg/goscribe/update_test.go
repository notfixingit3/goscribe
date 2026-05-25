package goscribe

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/house/goscribe/internal/docs"
)

func initGitRepo(t *testing.T, dir string) *git.Repository {
	t.Helper()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("git init: %v", err)
	}
	return repo
}

func commitFile(t *testing.T, repo *git.Repository, dir, filename, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("get worktree: %v", err)
	}
	if _, err := w.Add(filename); err != nil {
		t.Fatalf("git add: %v", err)
	}
	commit, err := w.Commit("add "+filename, &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	if err != nil {
		t.Fatalf("git commit: %v", err)
	}
	return commit.String()
}

func TestUpdate_NoChanges(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	repo := initGitRepo(t, srcDir)
	commit := commitFile(t, repo, srcDir, "main.go", "package main\n")

	if err := docs.SaveCommitState(srcDir, commit); err != nil {
		t.Fatalf("save state: %v", err)
	}

	ctx := context.Background()
	result, err := client.Update(ctx, srcDir, UpdateOptions{})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !result.NoChanges {
		t.Fatal("expected NoChanges to be true")
	}
	if result.FilesUpdated != 0 {
		t.Fatalf("expected FilesUpdated 0, got %d", result.FilesUpdated)
	}
	if result.FromCommit != commit {
		t.Fatalf("expected FromCommit %s, got %s", commit, result.FromCommit)
	}
	if result.ToCommit != commit {
		t.Fatalf("expected ToCommit %s, got %s", commit, result.ToCommit)
	}
}

func TestUpdate_WithChangedFiles(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	repo := initGitRepo(t, srcDir)
	commit1 := commitFile(t, repo, srcDir, "main.go", "package main\n")

	if err := docs.SaveCommitState(srcDir, commit1); err != nil {
		t.Fatalf("save state: %v", err)
	}

	commit2 := commitFile(t, repo, srcDir, "main.go", "package main\n\nfunc main() {}\n")

	ctx := context.Background()
	result, err := client.Update(ctx, srcDir, UpdateOptions{})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result.NoChanges {
		t.Fatal("expected NoChanges to be false")
	}
	if result.FilesUpdated == 0 {
		t.Fatal("expected FilesUpdated > 0")
	}
	if result.FromCommit != commit1 {
		t.Fatalf("expected FromCommit %s, got %s", commit1, result.FromCommit)
	}
	if result.ToCommit != commit2 {
		t.Fatalf("expected ToCommit %s, got %s", commit2, result.ToCommit)
	}
	if !result.StateSaved {
		t.Fatal("expected StateSaved to be true")
	}
}

func TestUpdate_MissingStateFile(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	repo := initGitRepo(t, srcDir)
	commitFile(t, repo, srcDir, "main.go", "package main\n")

	ctx := context.Background()
	_, err = client.Update(ctx, srcDir, UpdateOptions{})
	if err == nil {
		t.Fatal("expected error when state file is missing")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", gerr.Kind)
	}
}

func TestUpdate_NotGitRepository(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()

	ctx := context.Background()
	_, err = client.Update(ctx, srcDir, UpdateOptions{})
	if err == nil {
		t.Fatal("expected error when source is not a git repo")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrNotGitRepository) {
		t.Fatalf("expected ErrNotGitRepository, got %v", gerr.Kind)
	}
}

func TestUpdate_StateFileHandling(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	repo := initGitRepo(t, srcDir)
	commit1 := commitFile(t, repo, srcDir, "main.go", "package main\n")

	if err := docs.SaveCommitState(srcDir, commit1); err != nil {
		t.Fatalf("save state: %v", err)
	}

	_ = commitFile(t, repo, srcDir, "main.go", "package main\n\nfunc main() {}\n")

	ctx := context.Background()
	result, err := client.Update(ctx, srcDir, UpdateOptions{SkipStateSave: true})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result.StateSaved {
		t.Fatal("expected StateSaved to be false when SkipStateSave is true")
	}

	statePath := filepath.Join(srcDir, ".goscribe-state")
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}
	if string(data) == "" {
		t.Fatal("expected state file to still contain original commit")
	}
}

func TestUpdate_ExplicitCommits(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	repo := initGitRepo(t, srcDir)
	commit1 := commitFile(t, repo, srcDir, "main.go", "package main\n")
	commit2 := commitFile(t, repo, srcDir, "main.go", "package main\n\nfunc main() {}\n")

	if err := docs.SaveCommitState(srcDir, commit1); err != nil {
		t.Fatalf("save state: %v", err)
	}

	ctx := context.Background()
	result, err := client.Update(ctx, srcDir, UpdateOptions{
		FromCommit:    commit1,
		ToCommit:      commit2,
		SkipStateSave: true,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result.NoChanges {
		t.Fatal("expected NoChanges to be false")
	}
	if result.FromCommit != commit1 {
		t.Fatalf("expected FromCommit %s, got %s", commit1, result.FromCommit)
	}
	if result.ToCommit != commit2 {
		t.Fatalf("expected ToCommit %s, got %s", commit2, result.ToCommit)
	}
}

func TestUpdate_ExplicitChangedFiles(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	repo := initGitRepo(t, srcDir)
	commit1 := commitFile(t, repo, srcDir, "main.go", "package main\n")
	_ = commitFile(t, repo, srcDir, "main.go", "package main\n\nfunc main() {}\n")

	if err := docs.SaveCommitState(srcDir, commit1); err != nil {
		t.Fatalf("save state: %v", err)
	}

	ctx := context.Background()
	result, err := client.Update(ctx, srcDir, UpdateOptions{
		ChangedFiles:  []string{"main.go"},
		SkipStateSave: true,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result.NoChanges {
		t.Fatal("expected NoChanges to be false")
	}
	if len(result.ChangedFiles) != 1 {
		t.Fatalf("expected 1 changed file, got %d", len(result.ChangedFiles))
	}
	if result.ChangedFiles[0] != "main.go" {
		t.Fatalf("expected changed file 'main.go', got %q", result.ChangedFiles[0])
	}
}

func TestUpdate_InvalidStateFile(t *testing.T) {
	mp := newMockProvider("doc")
	client, err := NewClient(WithProvider(mp))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	srcDir := t.TempDir()
	repo := initGitRepo(t, srcDir)
	commitFile(t, repo, srcDir, "main.go", "package main\n")

	statePath := filepath.Join(srcDir, ".goscribe-state")
	if err := os.WriteFile(statePath, []byte("invalid-json"), 0600); err != nil {
		t.Fatalf("write invalid state: %v", err)
	}

	ctx := context.Background()
	_, err = client.Update(ctx, srcDir, UpdateOptions{})
	if err == nil {
		t.Fatal("expected error for invalid state file")
	}
	var gerr *Error
	if !errors.As(err, &gerr) {
		t.Fatalf("expected *goscribe.Error, got %T", err)
	}
	if !errors.Is(gerr, ErrStateInvalid) && !errors.Is(gerr, ErrUpdateFailed) {
		t.Fatalf("expected ErrStateInvalid or ErrUpdateFailed, got %v", gerr.Kind)
	}
}
