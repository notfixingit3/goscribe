package wizard

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/house/goscribe/internal/docs"
)

// TestModelInit verifies NewModel() initializes with the correct default state.
func TestModelInit(t *testing.T) {
	m := NewModel()

	if m.state != StateSelectOperation {
		t.Errorf("expected initial state StateSelectOperation (%d), got %d", StateSelectOperation, m.state)
	}

	if m.cancelled {
		t.Error("expected cancelled to be false on initialization")
	}

	if m.err != nil {
		t.Errorf("expected nil error on initialization, got %v", m.err)
	}

	// Operation list should have 2 items.
	if got := len(m.operationList.Items()); got != 2 {
		t.Errorf("expected operation list to have 2 items, got %d", got)
	}

	// Init should return a non-nil command (textinput.Blink).
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init() to return a non-nil command")
	}
}

// TestProfileListPopulated verifies the profile list matches docs.RegisteredProfiles().
func TestProfileListPopulated(t *testing.T) {
	m := NewModel()

	expectedProfiles := docs.RegisteredProfiles()
	profileItems := m.profileList.Items()

	if len(profileItems) != len(expectedProfiles) {
		t.Fatalf("expected %d profiles in list, got %d", len(expectedProfiles), len(profileItems))
	}

	for i, item := range profileItems {
		pi, ok := item.(profileItem)
		if !ok {
			t.Fatalf("item %d is not a profileItem", i)
		}
		if pi.name != "" && pi.name != expectedProfiles[i] {
			t.Errorf("expected profile %q at index %d, got %q", expectedProfiles[i], i, pi.name)
		}
	}
}

// TestTemplateNone verifies the template list includes "<none>" as the first item
// and has the correct total count.
func TestTemplateNone(t *testing.T) {
	m := NewModel()

	templateItems := m.templateList.Items()
	expectedCount := len(docs.RegisteredTemplates()) + 1 // +1 for "<none>"

	if len(templateItems) != expectedCount {
		t.Fatalf("expected %d templates in list (including <none>), got %d", expectedCount, len(templateItems))
	}

	if len(templateItems) == 0 {
		t.Fatal("template list is empty")
	}

	first, ok := templateItems[0].(templateItem)
	if !ok {
		t.Fatal("first item is not a templateItem")
	}
	if first.name != "<none>" {
		t.Errorf("expected first template item to be \"<none>\", got %q", first.name)
	}
}

// TestStepTransitions verifies each step advances to the next on enter key.
func TestStepTransitions(t *testing.T) {
	m := NewModel()

	// StateSelectOperation -> StateSelectProfile
	item, ok := m.operationList.SelectedItem().(operationItem)
	if !ok {
		t.Fatal("no operation selected by default")
	}
	selectedOp := item.name

	newModel, cmd := m.Update(tea.KeyPressMsg{Text: "enter"})
	if cmd != nil {
		t.Errorf("expected nil cmd after operation enter, got %v", cmd)
	}
	wm := newModel.(wizardModel)
	if wm.state != StateSelectProfile {
		t.Errorf("expected StateSelectProfile after operation enter, got %d", wm.state)
	}
	if wm.selectedOperation != selectedOp {
		t.Errorf("expected selectedOperation %q, got %q", selectedOp, wm.selectedOperation)
	}

	// StateSelectProfile -> StateSelectTemplate
	item2, ok := wm.profileList.SelectedItem().(profileItem)
	if !ok {
		t.Fatal("no profile selected by default")
	}
	selectedProfile := item2.name

	newModel2, cmd := wm.Update(tea.KeyPressMsg{Text: "enter"})
	if cmd != nil {
		t.Errorf("expected nil cmd after profile enter, got %v", cmd)
	}
	wm2 := newModel2.(wizardModel)
	if wm2.state != StateSelectTemplate {
		t.Errorf("expected StateSelectTemplate after profile enter, got %d", wm2.state)
	}
	if wm2.selectedProfile != selectedProfile {
		t.Errorf("expected selectedProfile %q, got %q", selectedProfile, wm2.selectedProfile)
	}

	// StateSelectTemplate -> StateInputPath
	item3, ok := wm2.templateList.SelectedItem().(templateItem)
	if !ok {
		t.Fatal("no template selected by default")
	}
	selectedTemplate := item3.name

	newModel3, cmd := wm2.Update(tea.KeyPressMsg{Text: "enter"})
	if cmd != nil {
		t.Errorf("expected nil cmd after template enter, got %v", cmd)
	}
	wm3 := newModel3.(wizardModel)
	if wm3.state != StateInputPath {
		t.Errorf("expected StateInputPath after template enter, got %d", wm3.state)
	}
	if wm3.selectedTemplate != selectedTemplate {
		t.Errorf("expected selectedTemplate %q, got %q", selectedTemplate, wm3.selectedTemplate)
	}

	// StateInputPath (empty input) -> StateInputOutput with default "."
	newModel4, cmd := wm3.Update(tea.KeyPressMsg{Text: "enter"})
	if cmd != nil {
		t.Errorf("expected nil cmd after path enter, got %v", cmd)
	}
	wm4 := newModel4.(wizardModel)
	if wm4.state != StateInputOutput {
		t.Errorf("expected StateInputOutput after path enter, got %d", wm4.state)
	}
	if wm4.sourcePath != "." {
		t.Errorf("expected sourcePath to default to \".\", got %q", wm4.sourcePath)
	}

	// StateInputOutput (empty input) -> StateConfirm with default "docs"
	newModel5, cmd := wm4.Update(tea.KeyPressMsg{Text: "enter"})
	if cmd != nil {
		t.Errorf("expected nil cmd after output enter, got %v", cmd)
	}
	wm5 := newModel5.(wizardModel)
	if wm5.state != StateConfirm {
		t.Errorf("expected StateConfirm after output enter, got %d", wm5.state)
	}
	if wm5.outputDir != "docs" {
		t.Errorf("expected outputDir to default to \"docs\", got %q", wm5.outputDir)
	}

	// StateConfirm + "y" -> StateDone + tea.Quit
	newModel6, cmd := wm5.Update(tea.KeyPressMsg{Text: "y"})
	wm6 := newModel6.(wizardModel)
	if wm6.state != StateDone {
		t.Errorf("expected StateDone after confirm, got %d", wm6.state)
	}
	if wm6.cancelled {
		t.Error("expected cancelled to be false on confirm")
	}
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd after confirm, got nil")
	}
}

// TestCancelOnEsc verifies ESC at any step sets cancelled=true and returns tea.Quit.
func TestCancelOnEsc(t *testing.T) {
	// Test ESC at initial state.
	m := NewModel()
	newModel, cmd := m.Update(tea.KeyPressMsg{Text: "esc"})
	wm := newModel.(wizardModel)
	if !wm.cancelled {
		t.Error("expected cancelled=true after esc")
	}
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd after esc, got nil")
	}

	// Test ESC at profile selection state.
	m2 := NewModel()
	m2.state = StateSelectProfile
	newModel2, cmd2 := m2.Update(tea.KeyPressMsg{Text: "esc"})
	wm2 := newModel2.(wizardModel)
	if !wm2.cancelled {
		t.Error("expected cancelled=true after esc at profile state")
	}
	if cmd2 == nil {
		t.Fatal("expected tea.Quit cmd after esc at profile state, got nil")
	}

	// Test ESC at template selection state.
	m3 := NewModel()
	m3.state = StateSelectTemplate
	newModel3, cmd3 := m3.Update(tea.KeyPressMsg{Text: "esc"})
	wm3 := newModel3.(wizardModel)
	if !wm3.cancelled {
		t.Error("expected cancelled=true after esc at template state")
	}
	if cmd3 == nil {
		t.Fatal("expected tea.Quit cmd after esc at template state, got nil")
	}

	// Test ESC at input path state.
	m4 := NewModel()
	m4.state = StateInputPath
	newModel4, cmd4 := m4.Update(tea.KeyPressMsg{Text: "esc"})
	wm4 := newModel4.(wizardModel)
	if !wm4.cancelled {
		t.Error("expected cancelled=true after esc at path input state")
	}
	if cmd4 == nil {
		t.Fatal("expected tea.Quit cmd after esc at path input state, got nil")
	}

	// Test ESC at input output state.
	m5 := NewModel()
	m5.state = StateInputOutput
	newModel5, cmd5 := m5.Update(tea.KeyPressMsg{Text: "esc"})
	wm5 := newModel5.(wizardModel)
	if !wm5.cancelled {
		t.Error("expected cancelled=true after esc at output input state")
	}
	if cmd5 == nil {
		t.Fatal("expected tea.Quit cmd after esc at output input state, got nil")
	}

	// Test ESC at confirm state.
	m6 := NewModel()
	m6.state = StateConfirm
	newModel6, cmd6 := m6.Update(tea.KeyPressMsg{Text: "esc"})
	wm6 := newModel6.(wizardModel)
	if !wm6.cancelled {
		t.Error("expected cancelled=true after esc at confirm state")
	}
	if cmd6 == nil {
		t.Fatal("expected tea.Quit cmd after esc at confirm state, got nil")
	}
}
