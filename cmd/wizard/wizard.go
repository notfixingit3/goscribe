package wizard

import (
	"errors"
	"os"

	tea "charm.land/bubbletea/v2"
	"golang.org/x/term"
)

var ErrCancelled = errors.New("wizard cancelled")

// ErrNoTTY is returned when RunWizard is called but stdin is not a terminal.
var ErrNoTTY = errors.New("wizard requires a terminal")

// RunWizard starts the interactive TUI wizard and returns the user's selections.
func RunWizard() (WizardResult, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return WizardResult{}, ErrNoTTY
	}
	m := NewModel()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return WizardResult{}, err
	}

	wm, ok := finalModel.(wizardModel)
	if !ok {
		return WizardResult{}, errors.New("unexpected model type")
	}

	if wm.cancelled {
		return WizardResult{}, ErrCancelled
	}

	return WizardResult{
		Operation:  wm.selectedOperation,
		Profile:    wm.selectedProfile,
		Template:   wm.selectedTemplate,
		SourcePath: wm.sourcePath,
		OutputDir:  wm.outputDir,
	}, nil
}
