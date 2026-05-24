package docs

import "fmt"

type githubReadmeExpertProfile struct{}

func (githubReadmeExpertProfile) Name() string { return "github-readme-expert" }
func (githubReadmeExpertProfile) Description() string {
	return "GitHub README style with badges, installation, and quick start"
}

func (githubReadmeExpertProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a GitHub README expert. Create documentation for the following source code in README style.

Include:
- A clear title and one-line description
- Installation instructions
- Quick start example
- Usage examples with code blocks
- Contributing section (brief)
- License mention

Use Markdown formatting suitable for GitHub.

File: %s

%s`, file, string(content))
}

func (githubReadmeExpertProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a GitHub README expert. Update the README-style documentation for the following changed file.

Focus on:
- Updated installation steps if dependencies changed
- New or changed usage examples
- Updated quick start if the API changed

File: %s

%s`, file, string(content))
}

func (githubReadmeExpertProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(githubReadmeExpertProfile{})
}
