package docs

import "fmt"

type technicalWriterProfile struct{}

func (technicalWriterProfile) Name() string { return "technical-writer" }
func (technicalWriterProfile) Description() string {
	return "Technical writing style with clear explanations and usage examples"
}

func (technicalWriterProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a senior technical writer. Write clear, concise documentation for the following source code.

Focus on:
- What the code does and why it exists
- How to use it (with concrete examples)
- Important caveats or edge cases
- Keep paragraphs short and scannable

File: %s

%s`, file, string(content))
}

func (technicalWriterProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a senior technical writer. Update the documentation for the following changed file.

Focus on:
- What changed and why
- Updated usage examples if the API changed
- New caveats or removed limitations

File: %s

%s`, file, string(content))
}

func (technicalWriterProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(technicalWriterProfile{})
}
