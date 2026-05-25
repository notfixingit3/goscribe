package wizard

import (
	"errors"
	"os"

	tea "charm.land/bubbletea/v2"
	"golang.org/x/term"
)

// ErrCanceled is returned when the user cancels the wizard with ESC or Ctrl+C.
var ErrCanceled = errors.New("wizard canceled")

// ErrNoTTY is returned when RunWizard is called but stdin is not a terminal.
var ErrNoTTY = errors.New("wizard requires a terminal")

// RunWizard starts the interactive TUI wizard and returns the user's selections.
func RunWizard() (Result, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return Result{}, ErrNoTTY
	}
	m := newModel()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return Result{}, err
	}

	wm, ok := finalModel.(wizardModel)
	if !ok {
		return Result{}, errors.New("unexpected model type")
	}

	if wm.canceled {
		return Result{}, ErrCanceled
	}

	return Result{
		Operation:  wm.selectedOperation,
		Profile:    wm.selectedProfile,
		Template:   wm.selectedTemplate,
		SourcePath: wm.sourcePath,
		OutputDir:  wm.outputDir,
	}, nil
}
