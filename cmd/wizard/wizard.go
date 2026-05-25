package wizard

import (
	"errors"

	tea "charm.land/bubbletea/v2"
)

var ErrCancelled = errors.New("wizard cancelled")

func RunWizard(profiles, templates []string) (WizardResult, error) {
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
