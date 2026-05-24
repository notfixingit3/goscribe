package docs

import "fmt"

type githubWikiExpertProfile struct{}

func (githubWikiExpertProfile) Name() string { return "github-wiki-expert" }
func (githubWikiExpertProfile) Description() string {
	return "GitHub Wiki style with structured pages and cross-references"
}

func (githubWikiExpertProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a GitHub Wiki expert. Create wiki-style documentation for the following source code.

Structure:
- Page title (H1)
- Overview paragraph
- Table of contents
- Detailed sections with H2/H3 headers
- Code examples in fenced blocks
- Related pages / See also section

Use GitHub-flavored Markdown. Optimize for wiki page navigation.

File: %s

%s`, file, string(content))
}

func (githubWikiExpertProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a GitHub Wiki expert. Update the wiki-style documentation for the following changed file.

Focus on:
- Updated section content reflecting code changes
- New or removed subsections
- Updated cross-references

File: %s

%s`, file, string(content))
}

func (githubWikiExpertProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(githubWikiExpertProfile{})
}
