// Package version manages the goscribe version string and related utilities.
package version

import (
	"fmt"
	"os/exec"
	"strings"
)

// Version is the current goscribe version, overridden at build time via ldflags.
var Version = "0.0.1"

var scoobyQuotes = []string{
	"Ruh-roh! Looks like we got a new version, Raggy!",
	"Scooby-Dooby-Doo! Version bump time!",
	"Zoinks! That's one spooky version update!",
	"Jinkies! The version just got smarter!",
	"Would you do it for a Scooby Snack? Version bumped!",
	"Ruh-roh-RAGGY! New version incoming!",
	"Scooby Dooby Doo! Where are you? In the new version!",
	"Puppy Power! Version updated!",
}

// Get returns the current version string.
func Get() string {
	return Version
}

// BumpPatch increments the patch version and updates the package-level Version.
func BumpPatch() (string, error) {
	parts := strings.Split(Version, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version format: %s", Version)
	}

	var patch int
	if _, err := fmt.Sscanf(parts[2], "%d", &patch); err != nil {
		return "", fmt.Errorf("parse patch version: %w", err)
	}

	patch++
	Version = fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch)
	return Version, nil
}

// GetScoobyQuote returns a Scooby-Doo quote selected by the current version hash.
func GetScoobyQuote() string {
	// In a real implementation, you'd use crypto/rand for randomness
	// For simplicity, we'll use a simple hash-based approach
	hash := 0
	for _, c := range Version {
		hash += int(c)
	}
	return scoobyQuotes[hash%len(scoobyQuotes)]
}

// GetGitVersion returns the git describe output for the current repository.
func GetGitVersion() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--always")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("get git version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
