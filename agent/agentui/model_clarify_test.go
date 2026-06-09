package agentui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestClarificationTabSkipsAnswered(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		selIdx:            0,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
		},
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	um := updated.(model)
	q := um.currentPendingQuestion()
	if q == nil {
		t.Fatal("expected a pending question")
	}
	if q.ID != "2" {
		t.Errorf("Tab should land on unanswered Q2, got %s", q.ID)
	}
}

func TestClarificationCtrlCFirstTimeSetsPending(t *testing.T) {
	m := model{
		done:              true,
		clarificationMode: true,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
		},
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	um := updated.(model)
	if !um.ctrlCPending {
		t.Error("first ctrl+c in clarification should set ctrlCPending=true")
	}
	if cmd == nil {
		t.Error("first ctrl+c should return a timer command")
	}
}

func TestClarificationOptionNavDown(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}, {Label: "B"}, {Label: "C"}}},
		},
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	um := updated.(model)
	if um.optionHighlightIdx != 1 {
		t.Errorf("down should move to index 1, got %d", um.optionHighlightIdx)
	}

	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyDown})
	um = updated.(model)
	if um.optionHighlightIdx != 2 {
		t.Errorf("down should move to index 2, got %d", um.optionHighlightIdx)
	}

	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyDown})
	um = updated.(model)
	if um.optionHighlightIdx != 3 {
		t.Errorf("down should move to virtual option index 3, got %d", um.optionHighlightIdx)
	}

	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyDown})
	um = updated.(model)
	if um.optionHighlightIdx != 0 {
		t.Errorf("down should wrap to index 0, got %d", um.optionHighlightIdx)
	}
}

func TestClarificationOptionNavUp(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}, {Label: "B"}, {Label: "C"}}},
		},
		optionHighlightIdx: 3,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	um := updated.(model)
	if um.optionHighlightIdx != 2 {
		t.Errorf("up should move to index 2, got %d", um.optionHighlightIdx)
	}

	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyUp})
	um = updated.(model)
	if um.optionHighlightIdx != 1 {
		t.Errorf("up should move to index 1, got %d", um.optionHighlightIdx)
	}

	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyUp})
	um = updated.(model)
	if um.optionHighlightIdx != 0 {
		t.Errorf("up should move to index 0, got %d", um.optionHighlightIdx)
	}

	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyUp})
	um = updated.(model)
	if um.optionHighlightIdx != 3 {
		t.Errorf("up should wrap to virtual option index 3, got %d", um.optionHighlightIdx)
	}
}

func TestClarificationOptionNavNoOptions(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Freeform?"},
		},
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	um := updated.(model)
	if um.optionHighlightIdx != 0 {
		t.Errorf("down on no-options should keep idx 0, got %d", um.optionHighlightIdx)
	}

	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyUp})
	um = updated.(model)
	if um.optionHighlightIdx != 0 {
		t.Errorf("up on no-options should keep idx 0, got %d", um.optionHighlightIdx)
	}
}

func TestClarificationEnterSubmitsHighlightedOption(t *testing.T) {
	vp := viewport.New(80, 20)
	m := &model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		logs:              []string{},
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "Card"}, {Label: "PayPal"}}},
		},
		optionHighlightIdx: 1,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(model)
	if um.questions[0].Answer != "PayPal" {
		t.Errorf("expected answer 'PayPal', got %q", um.questions[0].Answer)
	}
}

func TestClarificationEnterSubmitsTypedTextOverOption(t *testing.T) {
	input := textinput.New()
	input.SetValue("Custom answer")
	vp := viewport.New(80, 20)
	m := &model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             input,
		width:             80,
		logs:              []string{},
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "Card"}, {Label: "PayPal"}}},
		},
		optionHighlightIdx: 1,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(model)
	if um.questions[0].Answer != "Custom answer" {
		t.Errorf("expected answer 'Custom answer', got %q", um.questions[0].Answer)
	}
}

func TestClarificationEnterEmptyNoOptions(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		logs:              []string{},
		questions: []pendingQuestion{
			{ID: "1", Question: "Freeform?"},
		},
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(model)
	if um.questions[0].Answer != "" {
		t.Errorf("enter with empty input and no options should not submit, got answer %q", um.questions[0].Answer)
	}
}

func TestClarificationTabResetsOptionHighlight(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Options: []QuestionOption{{Label: "A"}, {Label: "B"}}, Answer: "done"},
			{ID: "2", Question: "Q2", Options: []QuestionOption{{Label: "X"}, {Label: "Y"}}},
		},
		optionHighlightIdx: 1,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	um := updated.(model)
	if um.optionHighlightIdx != 0 {
		t.Errorf("tab should reset optionHighlightIdx to 0, got %d", um.optionHighlightIdx)
	}
	if um.selIdx != 1 {
		t.Errorf("tab should move selIdx, got %d", um.selIdx)
	}
}

func TestClarificationShiftTab(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		selIdx:            1,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
			{ID: "2", Question: "Q2"},
		},
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	um := updated.(model)
	if um.selIdx != 0 {
		t.Errorf("shift+tab should move to previous question, got selIdx %d", um.selIdx)
	}

	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	um = updated.(model)
	if um.selIdx != 1 {
		t.Errorf("shift+tab should wrap, got selIdx %d", um.selIdx)
	}
}

func TestClarificationShiftTabResetsOptionHighlight(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		selIdx:            0,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Options: []QuestionOption{{Label: "A"}, {Label: "B"}}},
		},
		optionHighlightIdx: 1,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	um := updated.(model)
	if um.optionHighlightIdx != 0 {
		t.Errorf("shift+tab should reset optionHighlightIdx to 0, got %d", um.optionHighlightIdx)
	}
}

func TestClarificationEnterOnVirtualSubmitsTypedText(t *testing.T) {
	input := textinput.New()
	input.SetValue("Custom response")
	vp := viewport.New(80, 20)
	m := &model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             input,
		width:             80,
		logs:              []string{},
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}}},
		},
		optionHighlightIdx: 1,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(model)
	if um.questions[0].Answer != "Custom response" {
		t.Errorf("expected answer 'Custom response' from virtual option, got %q", um.questions[0].Answer)
	}
}

func TestClarificationEnterOnVirtualEmptyDoesNothing(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		logs:              []string{},
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}}},
		},
		optionHighlightIdx: 1,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(model)
	if um.questions[0].Answer != "" {
		t.Errorf("enter on virtual with empty input should not submit, got answer %q", um.questions[0].Answer)
	}
}

func TestClarificationTypingIgnoredOnRealOption(t *testing.T) {
	vp := viewport.New(80, 20)
	input := textinput.New()
	input.Focus()
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             input,
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}, {Label: "B"}}},
		},
		optionHighlightIdx: 0,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	um := updated.(model)
	if um.input.Value() != "" {
		t.Errorf("typing should be ignored when on real option, got input %q", um.input.Value())
	}
	if um.optionHighlightIdx != 0 {
		t.Errorf("option highlight should not change on typing, got %d", um.optionHighlightIdx)
	}
}

func TestClarificationTypingAllowedOnVirtualOption(t *testing.T) {
	vp := viewport.New(80, 20)
	input := textinput.New()
	input.Focus()
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             input,
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}}},
		},
		optionHighlightIdx: 1,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	um := updated.(model)
	if !strings.Contains(um.input.Value(), "x") {
		t.Errorf("typing should work when on virtual option, got input %q", um.input.Value())
	}
}
