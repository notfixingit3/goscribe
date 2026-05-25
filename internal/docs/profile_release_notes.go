package docs

import "fmt"

type releaseNotesProfile struct{}

func (releaseNotesProfile) Name() string { return "release-notes" }
func (releaseNotesProfile) Description() string {
	return "Release notes and changelog-style documentation"
}

func (releaseNotesProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a release notes writer for a software project. Document this file's role in release-note style.

Focus on:
- What the file provides and its purpose in the project
- What depends on it (callers, consumers, or downstream impact)
- Notable design choices or architectural decisions
- Breaking change potential (API changes, behavior changes, or removal risk)

Structure the output as:
- Summary
- Changes
- Migration Notes
- Compatibility

File: %s

%s`, file, string(content))
}

func (releaseNotesProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a release notes writer for a software project. Write changelog-style notes for the following changed file.

Focus on:
- What was added, changed, or removed
- Migration steps if the API changed
- Version compatibility notes

File: %s

%s`, file, string(content))
}

func (releaseNotesProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(releaseNotesProfile{})
}
