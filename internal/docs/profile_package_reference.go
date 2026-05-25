package docs

import "fmt"

type packageReferenceProfile struct{}

func (packageReferenceProfile) Name() string { return "package-reference" }
func (packageReferenceProfile) Description() string {
	return "Package-level reference documentation with exported API details"
}

func (packageReferenceProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are writing per-package reference documentation. Document this Go package — its purpose, all exported types and functions with signatures, usage context, relationships to other packages, import path, and key internal design notes.

Structure your output as follows:
- Package Overview
- Exported Types
- Exported Functions
- Usage Examples
- Related Packages

File: %s

%s`, file, string(content))
}

func (packageReferenceProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`Update the package reference documentation for the following changed file.

Focus on:
- New exports
- Removed or deprecated symbols
- Changed signatures
- Updated relationships to other packages

File: %s

%s`, file, string(content))
}

func (packageReferenceProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(packageReferenceProfile{})
}
