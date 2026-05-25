//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/house/goscribe/internal/ci"
)

// TestCIMode_Generate_InvalidPath verifies that generate --ci with an invalid
// path returns exit code 1 and produces structured error output.
func TestCIMode_Generate_InvalidPath(t *testing.T) {
	c := buildOnce(t)

	stdout, stderr, exitCode := c.RunCommand("generate", "--ci", "/nonexistent/path/that/does/not/exist")

	if exitCode != 1 {
		t.Fatalf("expected exit code 1 for invalid path, got %d", exitCode)
	}

	combined := stdout + stderr
	if combined == "" {
		t.Fatal("expected output (stdout or stderr) for invalid path, got none")
	}

	if !strings.Contains(combined, "path") && !strings.Contains(combined, "does not exist") {
		t.Errorf("expected error to mention path issue, got: %s", combined)
	}
}

// TestCIMode_Generate_TextOutput verifies that generate --ci produces
// structured text output and exits 0 on success (or non-zero on provider error).
func TestCIMode_Generate_TextOutput(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	tmpDir := SetupTempDir(t)

	stdout, stderr, exitCode := c.RunCommand("--config", configPath, "generate", "--ci", tmpDir)
	combined := stdout + stderr

	if exitCode == 0 {
		if !strings.Contains(combined, "Documentation generated successfully") {
			t.Errorf("expected success message in CI text output, got: %s", combined)
		}
	} else if exitCode == 1 {
		if combined == "" {
			t.Fatal("expected structured error output in CI mode, got none")
		}
	} else {
		t.Fatalf("unexpected exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
}

// TestCIMode_Generate_JSONOutput verifies that generate --ci --output-format json
// produces valid JSON output.
func TestCIMode_Generate_JSONOutput(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	tmpDir := SetupTempDir(t)

	stdout, stderr, exitCode := c.RunCommand(
		"--config", configPath,
		"generate", "--ci", "--output-format", "json",
		tmpDir,
	)
	combined := stdout + stderr

	if exitCode != 0 && exitCode != 1 {
		t.Fatalf("unexpected exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	jsonStr := extractJSON(combined)
	if jsonStr == "" {
		t.Fatalf("expected JSON output in CI mode, got: stdout=%q stderr=%q", stdout, stderr)
	}

	var result ci.Result
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, jsonStr)
	}

	if result.Command != "generate" {
		t.Errorf("expected command 'generate', got %q", result.Command)
	}

	if exitCode == 0 {
		if !result.Success {
			t.Errorf("expected success=true when exit code is 0, got success=%v", result.Success)
		}
		if result.FilesGenerated <= 0 {
			t.Errorf("expected files_generated > 0 on success, got %d", result.FilesGenerated)
		}
		if result.OutputDir == "" {
			t.Errorf("expected output_dir on success, got empty")
		}
	} else {
		if result.Success {
			t.Errorf("expected success=false when exit code is 1, got success=%v", result.Success)
		}
		if result.Error == "" {
			t.Errorf("expected error message on failure, got empty")
		}
	}
}

// TestCIMode_Generate_MarkdownOutput verifies that generate --ci --output-format markdown
// produces markdown output.
func TestCIMode_Generate_MarkdownOutput(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	tmpDir := SetupTempDir(t)

	stdout, stderr, exitCode := c.RunCommand(
		"--config", configPath,
		"generate", "--ci", "--output-format", "markdown",
		tmpDir,
	)
	combined := stdout + stderr

	if exitCode != 0 && exitCode != 1 {
		t.Fatalf("unexpected exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if combined == "" {
		t.Fatalf("expected markdown output in CI mode, got none")
	}

	if !strings.Contains(combined, "## GoScribe:") {
		t.Errorf("expected markdown header '## GoScribe:', got: %s", combined)
	}
	if !strings.Contains(combined, "**Status:**") {
		t.Errorf("expected markdown status line, got: %s", combined)
	}

	if exitCode == 0 {
		if !strings.Contains(combined, "**Status:** Success") {
			t.Errorf("expected success status in markdown, got: %s", combined)
		}
	} else {
		if !strings.Contains(combined, "**Status:** Failed") {
			t.Errorf("expected failed status in markdown, got: %s", combined)
		}
		if !strings.Contains(combined, "**Error:**") {
			t.Errorf("expected error line in markdown failure output, got: %s", combined)
		}
	}
}

// TestCIMode_Generate_GitHubOutput verifies that generate --ci --output-format github
// produces GitHub annotation format output.
func TestCIMode_Generate_GitHubOutput(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	tmpDir := SetupTempDir(t)

	stdout, stderr, exitCode := c.RunCommand(
		"--config", configPath,
		"generate", "--ci", "--output-format", "github",
		tmpDir,
	)
	combined := stdout + stderr

	if exitCode != 0 && exitCode != 1 {
		t.Fatalf("unexpected exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if combined == "" {
		t.Fatalf("expected GitHub annotation output in CI mode, got none")
	}

	if exitCode == 0 {
		if !strings.Contains(combined, "::notice::") {
			t.Errorf("expected ::notice:: in GitHub success output, got: %s", combined)
		}
	} else {
		if !strings.Contains(combined, "::error::") {
			t.Errorf("expected ::error:: in GitHub failure output, got: %s", combined)
		}
	}
}

// TestCIMode_Update_NoChanges verifies that update --ci with no changes returns
// exit code 2 and produces structured output indicating no changes.
func TestCIMode_Update_NoChanges(t *testing.T) {
	c := buildOnce(t)

	tmpDir := SetupTempDir(t)
	initGitRepo(t, tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test\n"), 0600); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	gitAdd(t, tmpDir, ".")
	gitCommit(t, tmpDir, "initial commit")

	commitHash := getCurrentCommit(t, tmpDir)
	statePath := filepath.Join(tmpDir, ".goscribe-state")
	stateContent := `{"last_commit": "` + commitHash + `"}`
	if err := os.WriteFile(statePath, []byte(stateContent), 0600); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	stdout, stderr, exitCode := c.RunCommand("update", "--ci", tmpDir)
	combined := stdout + stderr

	if exitCode != 2 {
		t.Fatalf("expected exit code 2 for no changes, got %d\nstdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if combined == "" {
		t.Fatal("expected output when no changes detected in CI mode, got none")
	}

	if !strings.Contains(combined, "No changes") && !strings.Contains(combined, "no changes") {
		t.Errorf("expected 'no changes' message, got: %s", combined)
	}
}

// TestCIMode_Update_WithChanges verifies that update --ci with changes returns
// exit code 0 and produces structured output.
func TestCIMode_Update_WithChanges(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	tmpDir := SetupTempDir(t)
	initGitRepo(t, tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test\n"), 0600); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	gitAdd(t, tmpDir, ".")
	gitCommit(t, tmpDir, "initial commit")

	commitHash := getCurrentCommit(t, tmpDir)
	statePath := filepath.Join(tmpDir, ".goscribe-state")
	stateContent := `{"last_commit": "` + commitHash + `"}`
	if err := os.WriteFile(statePath, []byte(stateContent), 0600); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "newfile.go"), []byte("package main\n\nfunc NewFunc() {}\n"), 0600); err != nil {
		t.Fatalf("failed to write new file: %v", err)
	}
	gitAdd(t, tmpDir, ".")
	gitCommit(t, tmpDir, "second commit")

	stdout, stderr, exitCode := c.RunCommand(
		"--config", configPath,
		"update", "--ci",
		tmpDir,
	)
	combined := stdout + stderr

	if exitCode == 0 {
		if combined == "" {
			t.Fatal("expected output when update succeeds in CI mode, got none")
		}
	} else if exitCode == 1 {
		if combined == "" {
			t.Fatal("expected structured error output in CI mode, got none")
		}
	} else {
		t.Fatalf("unexpected exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
}

// TestCIMode_Update_JSONNoChanges verifies that update --ci --output-format json
// with no changes returns exit code 2 and produces valid JSON.
func TestCIMode_Update_JSONNoChanges(t *testing.T) {
	c := buildOnce(t)

	tmpDir := SetupTempDir(t)
	initGitRepo(t, tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test\n"), 0600); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	gitAdd(t, tmpDir, ".")
	gitCommit(t, tmpDir, "initial commit")

	commitHash := getCurrentCommit(t, tmpDir)
	statePath := filepath.Join(tmpDir, ".goscribe-state")
	stateContent := `{"last_commit": "` + commitHash + `"}`
	if err := os.WriteFile(statePath, []byte(stateContent), 0600); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	stdout, stderr, exitCode := c.RunCommand(
		"update", "--ci", "--output-format", "json",
		tmpDir,
	)
	combined := stdout + stderr

	if exitCode != 2 {
		t.Fatalf("expected exit code 2 for no changes, got %d", exitCode)
	}

	jsonStr := extractJSON(combined)
	if jsonStr == "" {
		t.Fatalf("expected JSON output in CI mode, got: stdout=%q stderr=%q", stdout, stderr)
	}

	var result ci.Result
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, jsonStr)
	}

	if result.Command != "update" {
		t.Errorf("expected command 'update', got %q", result.Command)
	}
	if !result.Success {
		t.Errorf("expected success=true for no-changes case, got success=%v", result.Success)
	}
	if !result.NoChanges {
		t.Errorf("expected no_changes=true for no-changes case, got no_changes=%v", result.NoChanges)
	}
}

// TestCIMode_Generate_ExitCode1OnError verifies that generate --ci returns
// exit code 1 when generation fails (e.g., invalid path).
func TestCIMode_Generate_ExitCode1OnError(t *testing.T) {
	c := buildOnce(t)

	_, _, exitCode := c.RunCommand("generate", "--ci", "/this/path/definitely/does/not/exist")

	if exitCode != 1 {
		t.Fatalf("expected exit code 1 on error, got %d", exitCode)
	}
}

// TestCIMode_Generate_ExitCode0OnSuccess verifies that generate --ci returns
// exit code 0 when generation succeeds.
func TestCIMode_Generate_ExitCode0OnSuccess(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	tmpDir := SetupTempDir(t)

	_, _, exitCode := c.RunCommand("--config", configPath, "generate", "--ci", tmpDir)

	if exitCode == 0 {
		return
	}

	if exitCode != 1 {
		t.Fatalf("expected exit code 0 or 1, got %d", exitCode)
	}
}

// --- helpers ---

// extractJSON finds the first '{' in s and returns the substring from there
// to the end, which should be the JSON payload.
func extractJSON(s string) string {
	idx := strings.Index(s, "{")
	if idx == -1 {
		return ""
	}
	return s[idx:]
}

// gitAdd stages files in the given git repository.
func gitAdd(t *testing.T, dir, path string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "add", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\noutput: %s", err, output)
	}
}

// gitCommit creates a commit in the given git repository.
func gitCommit(t *testing.T, dir, message string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "commit", "-m", message)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\noutput: %s", err, output)
	}
}

// getCurrentCommit returns the current HEAD commit hash.
func getCurrentCommit(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse failed: %v\noutput: %s", err, output)
	}
	return strings.TrimSpace(string(output))
}
