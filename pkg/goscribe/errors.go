package goscribe

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure modes.
var (
	// ErrInvalidContext indicates the context was canceled or expired.
	ErrInvalidContext = errors.New("invalid context")

	// ErrInvalidPath indicates the source path is invalid or inaccessible.
	ErrInvalidPath = errors.New("invalid path")

	// ErrProviderNotConfigured indicates no AI provider is configured.
	ErrProviderNotConfigured = errors.New("provider not configured")

	// ErrUnsupportedProvider indicates the requested provider is not supported.
	ErrUnsupportedProvider = errors.New("unsupported provider")

	// ErrGenerationFailed indicates documentation generation failed.
	ErrGenerationFailed = errors.New("generation failed")

	// ErrUpdateFailed indicates documentation update failed.
	ErrUpdateFailed = errors.New("update failed")

	// ErrNotGitRepository indicates the source path is not a git repository.
	ErrNotGitRepository = errors.New("not a git repository")

	// ErrStateNotFound indicates the .goscribe-state file was not found.
	ErrStateNotFound = errors.New("state not found")

	// ErrStateInvalid indicates the .goscribe-state file is corrupted.
	ErrStateInvalid = errors.New("state invalid")
)

// Error wraps public goscribe operation failures.
type Error struct {
	Op   string
	Kind error
	Path string
	Err  error
}

func (e *Error) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("goscribe: %s on %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("goscribe: %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Is reports whether the error matches the target error or kind.
func (e *Error) Is(target error) bool {
	if e.Kind != nil && errors.Is(e.Kind, target) {
		return true
	}
	if e.Err != nil && errors.Is(e.Err, target) {
		return true
	}
	return false
}
