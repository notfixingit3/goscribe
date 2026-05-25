package docs

import (
	"strings"
	"testing"
)

func newProfileTestCases() []struct {
	name    string
	profile Profile
} {
	return []struct {
		name    string
		profile Profile
	}{
		{name: "release-notes", profile: releaseNotesProfile{}},
		{name: "architecture-overview", profile: architectureOverviewProfile{}},
		{name: "contributing-guide", profile: contributingGuideProfile{}},
		{name: "package-reference", profile: packageReferenceProfile{}},
	}
}

func TestNewProfiles_Registered(t *testing.T) {
	for _, tc := range newProfileTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := GetProfile(tc.name)
			if !ok {
				t.Fatalf("GetProfile(%q) returned false — profile not registered", tc.name)
			}
			if got == nil {
				t.Fatalf("GetProfile(%q) returned nil", tc.name)
			}
		})
	}
}

func TestNewProfiles_Names(t *testing.T) {
	for _, tc := range newProfileTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.profile.Name(); got != tc.name {
				t.Fatalf("Name() = %q, want %q", got, tc.name)
			}
		})
	}
}

func TestNewProfiles_Descriptions(t *testing.T) {
	for _, tc := range newProfileTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			desc := tc.profile.Description()
			if desc == "" {
				t.Fatal("Description() returned empty string")
			}
			if len(desc) < 10 {
				t.Fatalf("Description() too short (%d chars): %q", len(desc), desc)
			}
		})
	}
}

func TestNewProfiles_BuildPrompt(t *testing.T) {
	file := "main.go"
	content := []byte("package main\n\nfunc main() {}")

	for _, tc := range newProfileTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			prompt := tc.profile.BuildPrompt(file, content)
			if len(prompt) <= 100 {
				t.Fatalf("BuildPrompt() output is %d chars, expected >100", len(prompt))
			}
			if !strings.Contains(prompt, file) {
				t.Fatal("BuildPrompt() output should include the file name")
			}
			if !strings.Contains(prompt, string(content)) {
				t.Fatal("BuildPrompt() output should include the source content")
			}
		})
	}
}

func TestNewProfiles_BuildUpdatePrompt(t *testing.T) {
	file := "handler.go"
	content := []byte("func Handle() { return }")

	for _, tc := range newProfileTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			prompt := tc.profile.BuildUpdatePrompt(file, content)
			if len(prompt) <= 80 {
				t.Fatalf("BuildUpdatePrompt() output is %d chars, expected >80", len(prompt))
			}
			if !strings.Contains(prompt, file) {
				t.Fatal("BuildUpdatePrompt() output should include the file name")
			}
			if !strings.Contains(prompt, string(content)) {
				t.Fatal("BuildUpdatePrompt() output should include the source content")
			}
		})
	}
}

func TestNewProfiles_FormatOutput(t *testing.T) {
	input := "# Some markdown\n\nContent here."

	for _, tc := range newProfileTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.profile.FormatOutput(input); got != input {
				t.Fatalf("FormatOutput() = %q, want unmodified input", got)
			}
		})
	}
}

func TestTotalProfileCount(t *testing.T) {
	profiles := RegisteredProfiles()
	var foundNew int
	for _, name := range profiles {
		switch name {
		case "release-notes", "architecture-overview", "contributing-guide", "package-reference":
			foundNew++
		}
	}
	if foundNew != 4 {
		t.Fatalf("Expected 4 new profiles registered, found %d in %v", foundNew, profiles)
	}
	if len(profiles) != 11 {
		t.Fatalf("RegisteredProfiles() returned %d profiles, want 11. Got: %v", len(profiles), profiles)
	}
}
