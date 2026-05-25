package docs

import "fmt"

type contributingGuideProfile struct{}

func (contributingGuideProfile) Name() string { return "contributing-guide" }
func (contributingGuideProfile) Description() string {
	return "Contributing guide for open-source project documentation"
}

func (contributingGuideProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are creating a CONTRIBUTING.md for an open-source project. Document how to contribute to this code.

For the following source code, focus on:
- PR process (infer from git patterns in the code)
- Coding conventions visible in the code
- Testing instructions
- Development environment setup
- Code style guidelines
- Review expectations

Structure the output as:
1. Getting Started
2. Development Setup
3. Coding Conventions
4. Testing
5. PR Process
6. Code Review

File: %s

%s`, file, string(content))
}

func (contributingGuideProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`Update the contributing guide for the following changed file.

Focus on:
- Changed conventions
- New testing requirements
- Updated PR process
- New dependencies

File: %s

%s`, file, string(content))
}

func (contributingGuideProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(contributingGuideProfile{})
}
