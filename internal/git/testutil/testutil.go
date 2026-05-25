// Package testutil provides test helpers for creating temporary git repositories.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// MockRepo holds a temporary git repository for testing.
type MockRepo struct {
	Path string
	Repo *git.Repository
}

// NewMockRepo creates a temporary git repository with an initial commit.
// Call Cleanup when done to remove the temp directory.
func NewMockRepo() (*MockRepo, error) {
	dir, err := os.MkdirTemp("", "goscribe-test-repo-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("init repo: %w", err)
	}

	mr := &MockRepo{Path: dir, Repo: repo}

	if err := mr.commitFile("README.md", "# test\n"); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("initial commit: %w", err)
	}

	return mr, nil
}

// commitFile adds or updates a file and commits it.
func (m *MockRepo) commitFile(name, content string) error {
	fullPath := filepath.Join(m.Path, name)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil { // #nosec G301 -- test utility
		return fmt.Errorf("create dirs: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil { // #nosec G306 -- test utility
		return fmt.Errorf("write file: %w", err)
	}

	wt, err := m.Repo.Worktree()
	if err != nil {
		return fmt.Errorf("get worktree: %w", err)
	}

	_, err = wt.Add(name)
	if err != nil {
		return fmt.Errorf("add file: %w", err)
	}

	_, err = wt.Commit(name, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@test.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// AddFile adds a file and commits it, returning any error.
func (m *MockRepo) AddFile(name, content string) error {
	return m.commitFile(name, content)
}

// HeadCommit returns the current HEAD commit hash.
func (m *MockRepo) HeadCommit() (string, error) {
	head, err := m.Repo.Head()
	if err != nil {
		return "", fmt.Errorf("get head: %w", err)
	}
	return head.Hash().String(), nil
}

// Cleanup removes the temporary repository directory.
func (m *MockRepo) Cleanup() {
	_ = os.RemoveAll(m.Path) // #nosec G104 -- cleanup in test
}
