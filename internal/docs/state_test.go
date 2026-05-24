package docs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveCommitStateWritesJSON(t *testing.T) {
	dir := t.TempDir()
	commit := "abc123def456"

	if err := SaveCommitState(dir, commit); err != nil {
		t.Fatalf("SaveCommitState: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".goscribe-state"))
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}

	expected := `{
  "last_commit": "abc123def456"
}`
	if string(data) != expected {
		t.Errorf("state file content = %q, want %q", string(data), expected)
	}
}

func TestLoadCommitStateFromJSON(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "last_commit": "abc123def456"
}`
	if err := os.WriteFile(filepath.Join(dir, ".goscribe-state"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	commit, err := LoadCommitState(dir)
	if err != nil {
		t.Fatalf("LoadCommitState: %v", err)
	}
	if commit != "abc123def456" {
		t.Errorf("commit = %q, want %q", commit, "abc123def456")
	}
}

func TestLoadCommitStateLegacyRawString(t *testing.T) {
	dir := t.TempDir()
	rawCommit := "abc123def456"
	if err := os.WriteFile(filepath.Join(dir, ".goscribe-state"), []byte(rawCommit), 0644); err != nil {
		t.Fatal(err)
	}

	commit, err := LoadCommitState(dir)
	if err != nil {
		t.Fatalf("LoadCommitState: %v", err)
	}
	if commit != rawCommit {
		t.Errorf("commit = %q, want %q", commit, rawCommit)
	}
}

func TestLoadCommitStateMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadCommitState(dir)
	if err == nil {
		t.Error("expected error for missing state file")
	}
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	commit := "deadbeefcafe"

	if err := SaveCommitState(dir, commit); err != nil {
		t.Fatalf("SaveCommitState: %v", err)
	}

	loaded, err := LoadCommitState(dir)
	if err != nil {
		t.Fatalf("LoadCommitState: %v", err)
	}
	if loaded != commit {
		t.Errorf("round-trip: got %q, want %q", loaded, commit)
	}
}
