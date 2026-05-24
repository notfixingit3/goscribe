package docs

import "fmt"

type softwareDocumenterProfile struct{}

func (softwareDocumenterProfile) Name() string { return DefaultProfileName }
func (softwareDocumenterProfile) Description() string {
	return "General software documentation with examples and explanations"
}

func (softwareDocumenterProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`Generate comprehensive documentation for the following source code.
Include real examples and explain why they matter.

File: %s

%s`, file, string(content))
}

func (softwareDocumenterProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`Update the documentation for the following changed file.
Focus on what changed and why it matters.

File: %s

%s`, file, string(content))
}

func (softwareDocumenterProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(softwareDocumenterProfile{})
}
