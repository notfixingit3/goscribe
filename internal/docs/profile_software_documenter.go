package docs

import "fmt"

type softwareDocumenterProfile struct{}

func (softwareDocumenterProfile) Name() string { return DefaultProfileName }
func (softwareDocumenterProfile) Description() string {
	return "General software documentation with examples and explanations"
}

func (softwareDocumenterProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a senior software documentation engineer. Write comprehensive documentation for the following source code.

Focus on:
- What the code does and why it exists
- How to use it (with real-world examples)
- Important edge cases, gotchas, or limitations
- Keep paragraphs short and scannable

Structure the documentation as:
1. Overview: what this code is and its purpose
2. Installation / Setup (if applicable)
3. API / Usage: how to call or use the code
4. Examples: concrete, runnable examples with expected output
5. Edge Cases: known limitations, error conditions, or special behavior

Tone: Professional, clear, and helpful. Assume the reader is competent but new to this codebase.

File: %s

%s`, file, string(content))
}

func (softwareDocumenterProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a senior software documentation engineer updating existing documentation for the following changed file.

Focus on:
- What changed and why it changed
- Updated examples that reflect the new code
- New or removed edge cases, limitations, or behavior
- Preserve any still-accurate sections from the existing docs

Structure the updated documentation as:
1. Overview: revised purpose if it changed
2. Installation / Setup (if applicable)
3. API / Usage: updated signatures or behavior
4. Examples: updated or new concrete examples
5. Edge Cases: any new or removed limitations

Tone: Professional, clear, and helpful. Assume the reader is competent but new to this codebase.

File: %s

%s`, file, string(content))
}

func (softwareDocumenterProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(softwareDocumenterProfile{})
}
