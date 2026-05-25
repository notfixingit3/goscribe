//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var cli *CLI

// TestMain builds the goscribe binary once for all integration tests.
func TestMain(m *testing.M) {
	// Use a synthetic testing.T for build — TestMain doesn't have one.
	// We handle this by building in the first test that needs it instead.
	os.Exit(m.Run())
}

// buildOnce ensures the CLI binary is built exactly once.
func buildOnce(t *testing.T) *CLI {
	t.Helper()
	if cli == nil {
		cli = BuildBinary(t)
	}
	return cli
}

func TestVersionCommand(t *testing.T) {
	c := buildOnce(t)

	stdout, stderr, exitCode := c.RunCommand("version")

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "goscribe version") {
		t.Errorf("expected stdout to contain 'goscribe version', got: %s", stdout)
	}
	// Verify a version number is present (semver pattern: X.Y.Z)
	versionPattern := regexp.MustCompile(`\d+\.\d+\.\d+`)
	if !versionPattern.MatchString(stdout) {
		t.Errorf("expected stdout to contain a version number (X.Y.Z), got: %s", stdout)
	}
}

func TestGenerateCommand_NoProvider(t *testing.T) {
	c := buildOnce(t)

	tmpDir := SetupTempDir(t)
	stdout, stderr, exitCode := c.RunCommand("generate", tmpDir)

	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code, got %d", exitCode)
	}

	combined := stdout + stderr
	if !strings.Contains(strings.ToLower(combined), "provider") {
		t.Fatalf("expected provider error, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestGenerateCommand_WithConfig(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	tmpDir := SetupTempDir(t)
	stdout, stderr, exitCode := c.RunCommand("--config", configPath, "generate", tmpDir)
	combined := stdout + stderr

	if exitCode == 0 {
		if stdout == "" {
			t.Fatalf("expected output on success, got stdout=%q stderr=%q", stdout, stderr)
		}
		return
	}

	if !strings.Contains(combined, "localhost") && !strings.Contains(strings.ToLower(combined), "ollama") && !strings.Contains(strings.ToLower(combined), "connection refused") {
		t.Fatalf("expected config-based provider error, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestGenerateCommand_WithInvalidTemplate(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	tmpDir := SetupTempDir(t)
	stdout, stderr, exitCode := c.RunCommand("--config", configPath, "generate", "--template", "nonexistent", tmpDir)

	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code for invalid template, got %d", exitCode)
	}

	combined := stdout + stderr
	if !strings.Contains(strings.ToLower(combined), "unknown template") {
		t.Fatalf("expected unknown template error, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestGenerateCommand_WithProfileAndTemplate(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	tmpDir := SetupTempDir(t)
	stdout, stderr, exitCode := c.RunCommand(
		"--config", configPath,
		"generate",
		"--profile", "software-documenter",
		"--template", "elegant",
		tmpDir,
	)

	if exitCode != 0 && exitCode != 1 {
		t.Fatalf("expected exit code 0 or 1, got %d; stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	combined := stdout + stderr
	if strings.Contains(strings.ToLower(combined), "unknown profile") || strings.Contains(strings.ToLower(combined), "unknown template") {
		t.Fatalf("expected profile/template flags to be accepted, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestUpdateCommand_NoState(t *testing.T) {
	c := buildOnce(t)

	tmpDir := SetupTempDir(t)
	initGitRepo(t, tmpDir)

	stdout, stderr, exitCode := c.RunCommand("update", tmpDir)
	_ = stdout

	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code when no state file exists, got %d", exitCode)
	}

	combined := stdout + stderr
	if !strings.Contains(combined, ".goscribe-state") {
		t.Fatalf("expected error to mention state file, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestUpdateCommand_WithState(t *testing.T) {
	c := buildOnce(t)

	tmpDir := SetupTempDir(t)
	initGitRepo(t, tmpDir)
	statePath := filepath.Join(tmpDir, ".goscribe-state")
	stateContent := `{"last_commit": "abc123"}`
	if err := os.WriteFile(statePath, []byte(stateContent), 0600); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	stdout, stderr, exitCode := c.RunCommand("update", tmpDir)
	_ = stderr
	_ = exitCode

	if exitCode == 0 {
		// If it succeeds, great. If it fails because commit doesn't exist, that's also expected.
		return
	}

	combined := stdout + stderr
	if combined == "" {
		t.Fatalf("expected some output (stdout or stderr) when state commit is invalid, got none. exitCode=%d", exitCode)
	}
}

func TestProviderListCommand(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	stdout, stderr, exitCode := c.RunCommand("--config", configPath, "provider", "list")

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "ollama") {
		t.Errorf("expected stdout to contain 'ollama', got: %s", stdout)
	}
}

func TestProviderAddCommand(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	stdout, stderr, exitCode := c.RunCommand(
		"--config", configPath,
		"provider", "add", "openai",
		"--url", "https://api.openai.com",
		"--model", "gpt-4",
		"--key", "sk-test-key",
	)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "openai") {
		t.Errorf("expected stdout to contain 'openai', got: %s", stdout)
	}
}

func TestProviderRemoveCommand(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	stdout, stderr, exitCode := c.RunCommand(
		"--config", configPath,
		"provider", "remove", "ollama",
	)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "ollama") {
		t.Errorf("expected stdout to contain 'ollama', got: %s", stdout)
	}
}

func TestProviderAddCommand_MissingFlags(t *testing.T) {
	c := buildOnce(t)

	configPath := c.SetupTempConfig(t)
	stdout, stderr, exitCode := c.RunCommand(
		"--config", configPath,
		"provider", "add", "ollama",
		"--url", "http://localhost:8080",
		// Intentionally omit --model to test error handling
	)

	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code when --model is missing, got %d", exitCode)
	}

	combined := stdout + stderr
	if !strings.Contains(combined, "model") {
		t.Errorf("expected error to mention missing --model flag, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestBumpCommand(t *testing.T) {
	c := buildOnce(t)

	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# temp repo\n"), 0600); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}
	if output, err := exec.Command("git", "-C", tmpDir, "add", ".").CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\noutput: %s", err, output)
	}
	if output, err := exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\noutput: %s", err, output)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# temp repo\n\nchanged\n"), 0600); err != nil {
		t.Fatalf("failed to dirty working tree: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir to temp repo: %v", err)
	}

	stdout, stderr, exitCode := c.RunCommand("bump")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	combined := stdout + stderr
	for _, want := range []string{"Version bumped to", "Commit message: Bump version to", "Scooby quote:"} {
		if !strings.Contains(combined, want) {
			t.Fatalf("expected output to contain %q, got stdout=%q stderr=%q", want, stdout, stderr)
		}
	}
}

func TestTagCommand(t *testing.T) {
	c := buildOnce(t)

	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# temp repo\n"), 0600); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}
	if output, err := exec.Command("git", "-C", tmpDir, "add", ".").CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\noutput: %s", err, output)
	}
	if output, err := exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\noutput: %s", err, output)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir to temp repo: %v", err)
	}

	stdout, stderr, exitCode := c.RunCommand("tag")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if !strings.Contains(stdout, "Tagged 0.0.1") {
		t.Fatalf("expected tag output, got stdout=%q stderr=%q", stdout, stderr)
	}

	tagOut, err := exec.Command("git", "-C", tmpDir, "tag", "--list", "0.0.1").CombinedOutput()
	if err != nil {
		t.Fatalf("git tag --list failed: %v\noutput: %s", err, tagOut)
	}
	if strings.TrimSpace(string(tagOut)) != "0.0.1" {
		t.Fatalf("expected tag 0.0.1 to exist, got %q", string(tagOut))
	}
}

func TestHelpOutput(t *testing.T) {
	c := buildOnce(t)

	stdout, stderr, exitCode := c.RunCommand("--help")

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	commands := []string{"generate", "update", "provider", "version", "bump", "tag"}
	for _, cmd := range commands {
		if !strings.Contains(stdout, cmd) {
			t.Errorf("expected help output to contain '%s', got: %s", cmd, stdout)
		}
	}

	flags := []string{"--config", "--provider", "--model", "--output", "--verbose", "--timeout", "--retries"}
	for _, flag := range flags {
		if !strings.Contains(stdout, flag) {
			t.Errorf("expected help output to contain '%s', got: %s", flag, stdout)
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	c := buildOnce(t)

	stdout, stderr, exitCode := c.RunCommand("nonexistent")

	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code for unknown command, got %d", exitCode)
	}

	if !strings.Contains(stderr, "unknown command") {
		t.Errorf("expected stderr to contain 'unknown command', got: %s", stderr)
	}

	if stdout != "" {
		t.Errorf("expected empty stdout for unknown command, got: %s", stdout)
	}
}

func TestInvalidFlag(t *testing.T) {
	c := buildOnce(t)

	stdout, stderr, exitCode := c.RunCommand("--invalid-flag")

	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code for invalid flag, got %d", exitCode)
	}

	if !strings.Contains(stderr, "unknown flag") {
		t.Errorf("expected stderr to contain 'unknown flag', got: %s", stderr)
	}

	if stdout != "" {
		t.Errorf("expected empty stdout for invalid flag, got: %s", stdout)
	}
}
