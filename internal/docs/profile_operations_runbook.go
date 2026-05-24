package docs

import "fmt"

type operationsRunbookProfile struct{}

func (operationsRunbookProfile) Name() string { return "operations-runbook" }
func (operationsRunbookProfile) Description() string {
	return "Operations runbook with deployment, monitoring, and troubleshooting"
}

func (operationsRunbookProfile) BuildPrompt(file string, content []byte) string {
	return fmt.Sprintf(`You are an SRE writing an operations runbook. Document the following source code from an operations perspective.

Include:
- What this component does in production
- How to deploy or restart it
- Health checks and monitoring points
- Common failure modes and how to detect them
- Troubleshooting steps
- Dependencies and upstream/downstream services
- Resource requirements (if inferable)

Write for an on-call engineer who needs to fix things at 3 AM.

File: %s

%s`, file, string(content))
}

func (operationsRunbookProfile) BuildUpdatePrompt(file string, content []byte) string {
	return fmt.Sprintf(`Update the operations runbook for the following changed file.

Focus on:
- New deployment procedures
- Updated monitoring or health checks
- New failure modes
- Changed dependencies
- Updated resource requirements

File: %s

%s`, file, string(content))
}

func (operationsRunbookProfile) FormatOutput(doc string) string { return doc }

func init() {
	RegisterProfile(operationsRunbookProfile{})
}
