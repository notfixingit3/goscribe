package wizard

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.operationList.SetSize(msg.Width, msg.Height)
		m.profileList.SetSize(msg.Width, msg.Height)
		m.templateList.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "esc" {
			m.cancelled = true
			m.state = StateDone
			return m, tea.Quit
		}

		switch m.state {
		case StateSelectOperation:
			if msg.String() == "enter" {
				if item, ok := m.operationList.SelectedItem().(operationItem); ok {
					m.selectedOperation = item.name
					m.state = StateSelectProfile
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.operationList, cmd = m.operationList.Update(msg)
			return m, cmd

		case StateSelectProfile:
			if msg.String() == "enter" {
				if item, ok := m.profileList.SelectedItem().(profileItem); ok {
					m.selectedProfile = item.name
					m.state = StateSelectTemplate
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.profileList, cmd = m.profileList.Update(msg)
			return m, cmd

		case StateSelectTemplate:
			if msg.String() == "enter" {
				if item, ok := m.templateList.SelectedItem().(templateItem); ok {
					m.selectedTemplate = item.name
				}
				m.state = StateInputPath
				m.pathInput.Focus()
				return m, nil
			}
			var cmd tea.Cmd
			m.templateList, cmd = m.templateList.Update(msg)
			return m, cmd

		case StateInputPath:
			if msg.String() == "enter" {
				m.sourcePath = strings.TrimSpace(m.pathInput.Value())
				if m.sourcePath == "" {
					m.sourcePath = "."
				}
				m.state = StateInputOutput
				m.outputInput.Focus()
				return m, nil
			}
			var cmd tea.Cmd
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd

		case StateInputOutput:
			if msg.String() == "enter" {
				m.outputDir = strings.TrimSpace(m.outputInput.Value())
				if m.outputDir == "" {
					m.outputDir = "docs"
				}
				m.state = StateConfirm
				return m, nil
			}
			var cmd tea.Cmd
			m.outputInput, cmd = m.outputInput.Update(msg)
			return m, cmd

		case StateConfirm:
			switch msg.String() {
			case "y", "Y", "enter":
				m.state = StateDone
				return m, tea.Quit
			case "n", "N":
				m.state = StateInputPath
				m.pathInput.Focus()
				return m, nil
			}
			return m, nil
		}
	}

	return m, nil
}

func (m wizardModel) View() tea.View {
	var b strings.Builder

	switch m.state {
	case StateSelectOperation:
		b.WriteString("Select Operation:\n\n")
		b.WriteString(m.operationList.View())

	case StateSelectProfile:
		b.WriteString("Select a Documentation Profile:\n\n")
		b.WriteString(m.profileList.View())

	case StateSelectTemplate:
		b.WriteString("Select a Documentation Template:\n\n")
		b.WriteString(m.templateList.View())

	case StateInputPath:
		b.WriteString("Source Path:\n\n")
		b.WriteString(m.pathInput.View())
		b.WriteString("\n(press Enter to confirm)")

	case StateInputOutput:
		b.WriteString("Output Directory:\n\n")
		b.WriteString(m.outputInput.View())
		b.WriteString("\n(press Enter to confirm)")

	case StateConfirm:
		b.WriteString("Confirm your choices:\n\n")
		b.WriteString(fmt.Sprintf("Operation:  %s\n", m.selectedOperation))
		b.WriteString(fmt.Sprintf("Profile:    %s\n", m.selectedProfile))
		b.WriteString(fmt.Sprintf("Template:   %s\n", m.selectedTemplate))
		b.WriteString(fmt.Sprintf("Source:     %s\n", m.sourcePath))
		b.WriteString(fmt.Sprintf("Output:     %s\n\n", m.outputDir))
		b.WriteString("Press Y to confirm, N to edit, ESC to cancel")

	case StateDone:
		if m.cancelled {
			b.WriteString("Cancelled.")
		} else {
			b.WriteString("Done!")
		}
	}

	if m.state != StateDone {
		b.WriteString("\n\n(esc/ctrl+c to cancel)")
	}

	return tea.NewView(b.String())
}
