package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/house/goscribe/internal/git/testutil"
)

func TestOpenRepo(t *testing.T) {
	t.Run("valid git repo", func(t *testing.T) {
		mr, err := testutil.NewMockRepo()
		if err != nil {
			t.Fatalf("create mock repo: %v", err)
		}
		defer mr.Cleanup()

		repo, err := OpenRepo(mr.Path)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}
		if repo == nil {
			t.Fatal("expected non-nil repo")
		}
	})

	t.Run("non-git directory", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "goscribe-not-a-repo-*")
		if err != nil {
			t.Fatalf("create temp dir: %v", err)
		}
		defer os.RemoveAll(dir)

		repo, err := OpenRepo(dir)
		if err == nil {
			t.Fatal("expected error for non-git directory")
		}
		if repo != nil {
			t.Fatal("expected nil repo")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		repo, err := OpenRepo("/nonexistent/path/that/does/not/exist")
		if err == nil {
			t.Fatal("expected error for nonexistent path")
		}
		if repo != nil {
			t.Fatal("expected nil repo")
		}
	})
}

func TestGetCurrentCommit(t *testing.T) {
	t.Run("repo with commits", func(t *testing.T) {
		mr, err := testutil.NewMockRepo()
		if err != nil {
			t.Fatalf("create mock repo: %v", err)
		}
		defer mr.Cleanup()

		repo, err := OpenRepo(mr.Path)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}

		commit, err := repo.GetCurrentCommit()
		if err != nil {
			t.Fatalf("GetCurrentCommit failed: %v", err)
		}
		if commit == "" {
			t.Fatal("expected non-empty commit hash")
		}
		if len(commit) != 40 {
			t.Fatalf("expected 40-char hash, got %d: %s", len(commit), commit)
		}
	})

	t.Run("empty repo with no commits", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "goscribe-empty-repo-*")
		if err != nil {
			t.Fatalf("create temp dir: %v", err)
		}
		defer os.RemoveAll(dir)

		// Initialize repo but do not commit
		err = initEmptyRepo(dir)
		if err != nil {
			t.Fatalf("init empty repo: %v", err)
		}

		repo, err := OpenRepo(dir)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}

		_, err = repo.GetCurrentCommit()
		if err == nil {
			t.Fatal("expected error for repo with no commits")
		}
	})
}

func TestGetChangedFiles(t *testing.T) {
	t.Run("files changed between commits", func(t *testing.T) {
		mr, err := testutil.NewMockRepo()
		if err != nil {
			t.Fatalf("create mock repo: %v", err)
		}
		defer mr.Cleanup()

		firstCommit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get first commit: %v", err)
		}

		if err := mr.AddFile("main.go", "package main\n"); err != nil {
			t.Fatalf("add file: %v", err)
		}

		secondCommit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get second commit: %v", err)
		}

		repo, err := OpenRepo(mr.Path)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}

		files, err := repo.GetChangedFiles(firstCommit, secondCommit)
		if err != nil {
			t.Fatalf("GetChangedFiles failed: %v", err)
		}
		if len(files) == 0 {
			t.Fatal("expected changed files, got none")
		}
		if files[0] != "main.go" {
			t.Fatalf("expected main.go, got %s", files[0])
		}
	})

	t.Run("multiple files changed", func(t *testing.T) {
		mr, err := testutil.NewMockRepo()
		if err != nil {
			t.Fatalf("create mock repo: %v", err)
		}
		defer mr.Cleanup()

		firstCommit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get first commit: %v", err)
		}

		if err := mr.AddFile("a.go", "package a\n"); err != nil {
			t.Fatalf("add a.go: %v", err)
		}
		if err := mr.AddFile("b.go", "package b\n"); err != nil {
			t.Fatalf("add b.go: %v", err)
		}

		secondCommit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get second commit: %v", err)
		}

		repo, err := OpenRepo(mr.Path)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}

		files, err := repo.GetChangedFiles(firstCommit, secondCommit)
		if err != nil {
			t.Fatalf("GetChangedFiles failed: %v", err)
		}
		if len(files) != 2 {
			t.Fatalf("expected 2 changed files, got %d", len(files))
		}
	})

	t.Run("no changes between same commit", func(t *testing.T) {
		mr, err := testutil.NewMockRepo()
		if err != nil {
			t.Fatalf("create mock repo: %v", err)
		}
		defer mr.Cleanup()

		commit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get commit: %v", err)
		}

		repo, err := OpenRepo(mr.Path)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}

		files, err := repo.GetChangedFiles(commit, commit)
		if err != nil {
			t.Fatalf("GetChangedFiles failed: %v", err)
		}
		if len(files) != 0 {
			t.Fatalf("expected no changes, got %d files", len(files))
		}
	})

	t.Run("invalid from commit", func(t *testing.T) {
		mr, err := testutil.NewMockRepo()
		if err != nil {
			t.Fatalf("create mock repo: %v", err)
		}
		defer mr.Cleanup()

		commit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get commit: %v", err)
		}

		repo, err := OpenRepo(mr.Path)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}

		_, err = repo.GetChangedFiles("invalidhash", commit)
		if err == nil {
			t.Fatal("expected error for invalid from commit")
		}
	})

	t.Run("invalid to commit", func(t *testing.T) {
		mr, err := testutil.NewMockRepo()
		if err != nil {
			t.Fatalf("create mock repo: %v", err)
		}
		defer mr.Cleanup()

		commit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get commit: %v", err)
		}

		repo, err := OpenRepo(mr.Path)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}

		_, err = repo.GetChangedFiles(commit, "invalidhash")
		if err == nil {
			t.Fatal("expected error for invalid to commit")
		}
	})

	t.Run("file modified between commits", func(t *testing.T) {
		mr, err := testutil.NewMockRepo()
		if err != nil {
			t.Fatalf("create mock repo: %v", err)
		}
		defer mr.Cleanup()

		firstCommit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get first commit: %v", err)
		}

		if err := mr.AddFile("main.go", "package main\n"); err != nil {
			t.Fatalf("add main.go: %v", err)
		}
		if err := mr.AddFile("main.go", "package main\n\nfunc main() {}\n"); err != nil {
			t.Fatalf("modify main.go: %v", err)
		}

		secondCommit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get second commit: %v", err)
		}

		repo, err := OpenRepo(mr.Path)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}

		files, err := repo.GetChangedFiles(firstCommit, secondCommit)
		if err != nil {
			t.Fatalf("GetChangedFiles failed: %v", err)
		}
		if len(files) == 0 {
			t.Fatal("expected changed files, got none")
		}
		found := false
		for _, f := range files {
			if f == "main.go" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected main.go in changed files, got %v", files)
		}
	})

	t.Run("file deleted between commits", func(t *testing.T) {
		mr, err := testutil.NewMockRepo()
		if err != nil {
			t.Fatalf("create mock repo: %v", err)
		}
		defer mr.Cleanup()

		// Add a file to delete
		if err := mr.AddFile("delete_me.go", "package main\n"); err != nil {
			t.Fatalf("add delete_me.go: %v", err)
		}

		firstCommit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get first commit: %v", err)
		}

		// Delete the file
		fullPath := filepath.Join(mr.Path, "delete_me.go")
		if err := os.Remove(fullPath); err != nil {
			t.Fatalf("remove file: %v", err)
		}

		wt, err := mr.Repo.Worktree()
		if err != nil {
			t.Fatalf("get worktree: %v", err)
		}
		_, err = wt.Add("delete_me.go")
		if err != nil {
			t.Fatalf("stage deletion: %v", err)
		}
		_, err = wt.Commit("delete delete_me.go", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test",
				Email: "test@test.com",
				When:  time.Now(),
			},
		})
		if err != nil {
			t.Fatalf("commit deletion: %v", err)
		}

		secondCommit, err := mr.HeadCommit()
		if err != nil {
			t.Fatalf("get second commit: %v", err)
		}

		repo, err := OpenRepo(mr.Path)
		if err != nil {
			t.Fatalf("OpenRepo failed: %v", err)
		}

		files, err := repo.GetChangedFiles(firstCommit, secondCommit)
		if err != nil {
			t.Fatalf("GetChangedFiles failed: %v", err)
		}
		if len(files) == 0 {
			t.Fatal("expected changed files, got none")
		}
		found := false
		for _, f := range files {
			if f == "delete_me.go" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected delete_me.go in changed files, got %v", files)
		}
	})
}

// initEmptyRepo initializes a git repo without any commits.
func initEmptyRepo(dir string) error {
	_, err := git.PlainInit(dir, false)
	return err
}
