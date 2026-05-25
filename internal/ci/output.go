// Package ci provides structured output formatting for CI/CD environments.
package ci

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Result holds the outcome of a generate or update operation.
type Result struct {
	Success        bool     `json:"success"`
	Command        string   `json:"command"`
	FilesGenerated int      `json:"files_generated,omitempty"`
	FilesUpdated   int      `json:"files_updated,omitempty"`
	FilesChanged   []string `json:"files_changed,omitempty"`
	OutputDir      string   `json:"output_dir,omitempty"`
	Error          string   `json:"error,omitempty"`
	NoChanges      bool     `json:"no_changes,omitempty"`
}

// Format writes the result in the specified format to w.
// Supported formats: "json", "markdown", "github", "text".
func Format(w io.Writer, result Result, format string) {
	switch format {
	case "json":
		formatJSON(w, result)
	case "markdown":
		formatMarkdown(w, result)
	case "github":
		formatGitHub(w, result)
	default:
		formatText(w, result)
	}
}

func formatJSON(w io.Writer, result Result) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func formatMarkdown(w io.Writer, result Result) {
	if result.Success {
		_, _ = fmt.Fprintf(w, "## GoScribe: %s\n\n", title(result.Command))
		_, _ = fmt.Fprintf(w, "- **Status:** Success\n")
		if result.FilesGenerated > 0 {
			_, _ = fmt.Fprintf(w, "- **Files generated:** %d\n", result.FilesGenerated)
		}
		if result.FilesUpdated > 0 {
			_, _ = fmt.Fprintf(w, "- **Files updated:** %d\n", result.FilesUpdated)
		}
		if result.OutputDir != "" {
			_, _ = fmt.Fprintf(w, "- **Output directory:** %s\n", result.OutputDir)
		}
		if result.NoChanges {
			_, _ = fmt.Fprintf(w, "- **Note:** No changes detected\n")
		}
	} else {
		_, _ = fmt.Fprintf(w, "## GoScribe: %s\n\n", title(result.Command))
		_, _ = fmt.Fprintf(w, "- **Status:** Failed\n")
		_, _ = fmt.Fprintf(w, "- **Error:** %s\n", result.Error)
	}
}

func formatGitHub(w io.Writer, result Result) {
	if result.Success {
		if result.NoChanges {
			_, _ = fmt.Fprintf(w, "::notice::GoScribe %s: no changes detected\n", result.Command)
		} else {
			count := result.FilesGenerated + result.FilesUpdated
			_, _ = fmt.Fprintf(w, "::notice::GoScribe %s: success (%d files)\n", result.Command, count)
		}
	} else {
		_, _ = fmt.Fprintf(w, "::error::GoScribe %s: %s\n", result.Command, result.Error)
	}
}

func formatText(w io.Writer, result Result) {
	if result.Success {
		if result.NoChanges {
			_, _ = fmt.Fprintln(w, "No changes detected since last documentation generation.")
		} else if result.FilesGenerated > 0 {
			_, _ = fmt.Fprintf(w, "Documentation generated successfully in %s\n", result.OutputDir)
		} else if result.FilesUpdated > 0 {
			_, _ = fmt.Fprintf(w, "Documentation updated successfully (%d files changed)\n", result.FilesUpdated)
		}
	} else {
		_, _ = fmt.Fprintf(w, "Error: %s\n", result.Error)
	}
}
