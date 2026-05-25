// Package wizard provides an interactive TUI wizard for configuring goscribe runs.
package wizard

// WizardState represents the current step in the interactive wizard.
type WizardState int

const (
	// StateSelectOperation is the first step where the user chooses generate or update.
	StateSelectOperation WizardState = iota
	// StateSelectProfile is the step where the user selects a documentation profile.
	StateSelectProfile
	// StateSelectTemplate is the step where the user selects a documentation template.
	StateSelectTemplate
	// StateInputPath is the step where the user enters the source path.
	StateInputPath
	// StateInputOutput is the step where the user enters the output directory.
	StateInputOutput
	// StateConfirm is the step where the user confirms their choices before running.
	StateConfirm
	// StateDone is the final step after the wizard completes or is cancelled.
	StateDone
)
