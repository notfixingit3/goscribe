package docs

import "fmt"

type architectureOverviewProfile struct{}

func (architectureOverviewProfile) Name() string { return "architecture-overview" }
func (architectureOverviewProfile) Description() string {
	return "System architecture and component relationship documentation"
}

func (architectureOverviewProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a software architect documenting system design.

For the following source code, document its architectural role:

- Component Role: its place in the component hierarchy and what it does
- Dependencies: what it depends on and what depends on it
- Interfaces: interfaces it implements or consumes
- Design Decisions: key design decisions visible in the code
- Data Flow: data flow patterns through this component

File: %s

%s`, file, string(content))
}

func (architectureOverviewProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are a software architect documenting architectural changes.

For the following changed file, document the architectural changes:

- New dependencies added or removed
- Changed interfaces (implemented or consumed)
- Removed components or responsibilities
- Refactored patterns or design decisions
- Updated data flow

File: %s

%s`, file, string(content))
}

func (architectureOverviewProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(architectureOverviewProfile{})
}
