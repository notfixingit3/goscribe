//go:build integration

// Package integration provides end-to-end tests for the goscribe CLI binary.
package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// CLI holds the path to the built goscribe binary and provides helper methods
// for executing commands and setting up test environments.
type CLI struct {
	BinaryPath string
}

// BuildBinary compiles the goscribe binary to a temporary directory.
// It returns a CLI struct with the binary path. The caller should invoke
// cleanup when done.
func BuildBinary(t *testing.T) *CLI {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "goscribe-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir for binary: %v", err)
	}

	binName := "goscribe"
	if runtime.GOOS == "windows" {
		binName = "goscribe.exe"
	}
	binPath := filepath.Join(tmpDir, binName)

	// Determine the project root (two levels up from this file).
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file path")
	}
	projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	projectRoot, err = filepath.Abs(projectRoot)
	if err != nil {
		t.Fatalf("failed to resolve project root: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/goscribe")
	cmd.Dir = projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build goscribe binary: %v\noutput: %s", err, output)
	}

	return &CLI{BinaryPath: binPath}
}

// Cleanup removes the directory containing the built binary.
func (c *CLI) Cleanup() {
	if c.BinaryPath != "" {
		_ = os.RemoveAll(filepath.Dir(c.BinaryPath))
	}
}

// RunCommand executes the goscribe binary with the given arguments and
// returns stdout, stderr, and the exit code.
func (c *CLI) RunCommand(args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(c.BinaryPath, args...)

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
		return
	}
	exitCode = 0
	return
}

// RunCommandWithEnv executes the goscribe binary with custom environment
// variables appended to the current process environment.
func (c *CLI) RunCommandWithEnv(env []string, args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(c.BinaryPath, args...)
	cmd.Env = append(os.Environ(), env...)

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
		return
	}
	exitCode = 0
	return
}

// SetupTempConfig creates a temporary goscribe config file with a mock
// provider configuration. Returns the config file path and a cleanup function.
func (c *CLI) SetupTempConfig(t *testing.T) (configPath string) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath = filepath.Join(tmpDir, ".goscribe.yaml")

	content := `provider: ollama
model: test-model
output: docs
verbose: false
timeout: 30s
retries: 0
providers:
  - name: ollama
    url: http://localhost:11111
    model: test-model
    default: true
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return configPath
}

// SetupTempDir creates a temporary directory with sample Go source files
// suitable for testing generate/update commands. Returns the project directory
// path.
func SetupTempDir(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()

	// Create a minimal go.mod
	goMod := fmt.Sprintf("module example.com/testproject\n\ngo 1.26.0\n")
	if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goMod), 0600); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a sample main.go
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
`
	if err := os.WriteFile(filepath.Join(projectDir, "main.go"), []byte(mainGo), 0600); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create a sample lib.go
	libGo := `package main

// Greet returns a greeting for the given name.
func Greet(name string) string {
	return "hello " + name
}
`
	if err := os.WriteFile(filepath.Join(projectDir, "lib.go"), []byte(libGo), 0600); err != nil {
		t.Fatalf("failed to write lib.go: %v", err)
	}

	return projectDir
}

// initGitRepo initializes a git repository in the given directory.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to init git repo: %v\noutput: %s", err, output)
	}

	// Configure git user for commits
	exec.Command("git", "-C", dir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", dir, "config", "user.name", "Test User").Run()
}
