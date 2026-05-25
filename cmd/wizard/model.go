// Package wizard provides an interactive TUI wizard for configuring goscribe runs.
package wizard

import (
	"github.com/house/goscribe/internal/docs"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// WizardResult holds the final selections made by the user in the wizard.
type WizardResult struct {
	Operation  string
	Profile    string
	Template   string
	SourcePath string
	OutputDir  string
}

// operationItem is a list item representing a wizard operation.
type operationItem struct {
	name string
}

// Title returns the operation name for display in the list.
func (i operationItem) Title() string { return i.name }

// Description returns an empty string (operations have no description in the list).
func (i operationItem) Description() string { return "" }

// FilterValue returns the operation name for filtering.
func (i operationItem) FilterValue() string { return i.name }

// profileItem is a list item representing a documentation profile.
type profileItem struct {
	name string
}

// Title returns the profile name for display in the list.
func (i profileItem) Title() string { return i.name }

// Description returns an empty string (profiles have no description in the list).
func (i profileItem) Description() string { return "" }

// FilterValue returns the profile name for filtering.
func (i profileItem) FilterValue() string { return i.name }

// templateItem is a list item representing a documentation template.
type templateItem struct {
	name string
}

// Title returns the template name for display in the list.
func (i templateItem) Title() string { return i.name }

// Description returns an empty string (templates have no description in the list).
func (i templateItem) Description() string { return "" }

// FilterValue returns the template name for filtering.
func (i templateItem) FilterValue() string { return i.name }

// wizardModel is the bubbletea model for the interactive wizard.
type wizardModel struct {
	state             WizardState
	selectedOperation string
	selectedProfile   string
	selectedTemplate  string
	sourcePath        string
	outputDir         string
	operationList     list.Model
	profileList       list.Model
	templateList      list.Model
	pathInput         textinput.Model
	outputInput       textinput.Model
	width             int
	height            int
	cancelled         bool
	err               error
}

// NewModel creates and initializes a new wizardModel with default state and UI components.
func NewModel() wizardModel {
	// Profile list
	profileNames := docs.RegisteredProfiles()
	profileItems := make([]list.Item, len(profileNames))
	for i, name := range profileNames {
		profileItems[i] = profileItem{name: name}
	}
	profileList := list.New(profileItems, list.NewDefaultDelegate(), 0, 0)
	profileList.SetShowStatusBar(false)

	// Template list (includes "<none>" as the first option)
	templateNames := docs.RegisteredTemplates()
	templateItems := make([]list.Item, 0, len(templateNames)+1)
	templateItems = append(templateItems, templateItem{name: "<none>"})
	for _, name := range templateNames {
		templateItems = append(templateItems, templateItem{name: name})
	}
	templateList := list.New(templateItems, list.NewDefaultDelegate(), 0, 0)
	templateList.SetShowStatusBar(false)

	// Path input
	pathInput := textinput.New()
	pathInput.Placeholder = "."

	// Output input
	outputInput := textinput.New()
	outputInput.Placeholder = "docs"

	operationItems := []list.Item{
		operationItem{name: "Generate documentation"},
		operationItem{name: "Update documentation"},
	}
	operationList := list.New(operationItems, list.NewDefaultDelegate(), 0, 0)
	operationList.SetShowStatusBar(false)

	return wizardModel{
		state:         StateSelectOperation,
		operationList: operationList,
		profileList:   profileList,
		templateList:  templateList,
		pathInput:     pathInput,
		outputInput:   outputInput,
	}
}

// Init implements tea.Model. It returns a command to start the text input cursor blinking.
func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}
