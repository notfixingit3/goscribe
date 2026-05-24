package docs

import "fmt"

type developerOnboardingProfile struct{}

func (developerOnboardingProfile) Name() string { return "developer-onboarding" }
func (developerOnboardingProfile) Description() string {
	return "Onboarding guide for new developers joining the project"
}

func (developerOnboardingProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are writing onboarding documentation for a new developer joining this project.

For the following source code, explain:
- What this file does in the overall architecture
- Key concepts a new dev needs to understand
- How this code fits with other parts of the system
- Common patterns and conventions used
- Where to look for related code

Write in a welcoming, educational tone. Assume the reader is competent but unfamiliar with this codebase.

File: %s

%s`, file, string(content))
}

func (developerOnboardingProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`Update the onboarding documentation for the following changed file.

Focus on:
- Architecture changes that affect new developers
- New concepts or patterns introduced
- Updated relationships with other parts of the system
- Changed conventions

File: %s

%s`, file, string(content))
}

func (developerOnboardingProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(developerOnboardingProfile{})
}
