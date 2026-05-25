package version

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = "1.2.3"
	if got := Get(); got != "1.2.3" {
		t.Errorf("Get() = %q, want %q", got, "1.2.3")
	}
}

func TestBumpPatch(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	tests := []struct {
		name    string
		start   string
		want    string
		wantErr bool
	}{
		{"bump 0.0.1", "0.0.1", "0.0.2", false},
		{"bump 1.2.9", "1.2.9", "1.2.10", false},
		{"bump 0.0.0", "0.0.0", "0.0.1", false},
		{"invalid two parts", "0.0", "", true},
		{"invalid four parts", "0.0.0.1", "", true},
		{"invalid empty", "", "", true},
		{"invalid patch non-numeric", "0.0.abc", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.start
			got, err := BumpPatch()
			if (err != nil) != tt.wantErr {
				t.Errorf("BumpPatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("BumpPatch() = %q, want %q", got, tt.want)
			}
			if !tt.wantErr && Version != tt.want {
				t.Errorf("Version after BumpPatch() = %q, want %q", Version, tt.want)
			}
		})
	}
}

func TestBumpPatchSequential(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = "0.0.1"
	v1, err := BumpPatch()
	if err != nil {
		t.Fatalf("first BumpPatch: %v", err)
	}
	v2, err := BumpPatch()
	if err != nil {
		t.Fatalf("second BumpPatch: %v", err)
	}

	if v1 != "0.0.2" {
		t.Errorf("first BumpPatch = %q, want %q", v1, "0.0.2")
	}
	if v2 != "0.0.3" {
		t.Errorf("second BumpPatch = %q, want %q", v2, "0.0.3")
	}
}

func TestBumpPatchDoesNotAlterMajorMinor(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = "3.7.5"
	got, err := BumpPatch()
	if err != nil {
		t.Fatalf("BumpPatch: %v", err)
	}
	if !strings.HasPrefix(got, "3.7.") {
		t.Errorf("BumpPatch() = %q, want prefix %q", got, "3.7.")
	}
}

func TestGetScoobyQuote(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	// Verify we always get one of the known quotes
	seen := map[string]bool{}
	for _, q := range scoobyQuotes {
		seen[q] = true
	}

	// Test with several version strings to exercise different hash values
	versions := []string{
		"0.0.1", "0.0.2", "1.0.0", "9.9.9", "10.20.30",
	}
	for _, v := range versions {
		Version = v
		quote := GetScoobyQuote()
		if !seen[quote] {
			t.Errorf("GetScoobyQuote() with Version=%q returned unknown quote: %q", v, quote)
		}
	}
}

func TestGetScoobyQuoteReturnsNonEmpty(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = "0.0.1"
	quote := GetScoobyQuote()
	if quote == "" {
		t.Error("GetScoobyQuote() returned empty string")
	}
}

func TestGetScoobyQuoteDeterministic(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = "0.0.1"
	q1 := GetScoobyQuote()
	q2 := GetScoobyQuote()
	if q1 != q2 {
		t.Errorf("GetScoobyQuote() not deterministic for same version: %q vs %q", q1, q2)
	}
}

func TestScoobyQuotesList(t *testing.T) {
	if len(scoobyQuotes) == 0 {
		t.Error("scoobyQuotes list is empty")
	}
	for i, q := range scoobyQuotes {
		if strings.TrimSpace(q) == "" {
			t.Errorf("scoobyQuotes[%d] is empty or whitespace", i)
		}
	}
}

func TestGetGitVersion(t *testing.T) {
	// In a git repo this should succeed
	_, err := GetGitVersion()
	if err != nil {
		t.Logf("GetGitVersion() error (expected outside git repo or without tags): %v", err)
	}
}
