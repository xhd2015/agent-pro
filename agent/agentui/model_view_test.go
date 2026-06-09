package agentui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

func TestDoneViewShowsFollowUpPrompt(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:     true,
		viewport: vp,
		input:    textinput.New(),
		width:    80,
	}
	view := m.View()
	if !strings.Contains(view, "Enter to follow up") {
		t.Errorf("done view should contain 'Enter to follow up', got: %s", view)
	}
	if !strings.Contains(view, "Ctrl+C to exit") {
		t.Errorf("done view should contain 'Ctrl+C to exit', got: %s", view)
	}
}

func TestDoneViewDoesNotShowExitPrompt(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:     true,
		viewport: vp,
		input:    textinput.New(),
		width:    80,
	}
	view := m.View()
	if strings.Contains(view, "Press any key to exit") {
		t.Error("done view should no longer show 'Press any key to exit'")
	}
}

func TestNonDoneViewShowsWorking(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:     false,
		viewport: vp,
		input:    textinput.New(),
		width:    80,
	}
	view := m.View()
	if !strings.Contains(view, "Working") {
		t.Errorf("non-done view should show 'Working...', got: %s", view)
	}
}

func TestClarificationModeView(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "What platform?"},
			{ID: "2", Question: "Dark mode?", Options: []QuestionOption{{Label: "Yes"}, {Label: "No"}}},
		},
	}
	view := m.View()
	if !strings.Contains(view, "Clarification needed") {
		t.Errorf("clarification view should contain 'Clarification needed', got: %s", view)
	}
	if !strings.Contains(view, "What platform?") {
		t.Errorf("clarification view should show current question, got: %s", view)
	}
	if !strings.Contains(view, "Tab:next") {
		t.Errorf("clarification view should contain 'Tab:next', got: %s", view)
	}
}

func TestClarificationModeShowsOptions(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Payment?", Options: []QuestionOption{{Label: "Card"}, {Label: "PayPal"}}},
		},
	}
	view := m.View()
	if !strings.Contains(view, "Card") {
		t.Errorf("clarification view should show option 'Card', got: %s", view)
	}
	if !strings.Contains(view, "PayPal") {
		t.Errorf("clarification view should show option 'PayPal', got: %s", view)
	}
}

func TestClarificationShowsAnsweredCount(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
			{ID: "3", Question: "Q3"},
		},
	}
	view := m.View()
	if !strings.Contains(view, "1 of 3 answered") {
		t.Errorf("clarification view should show '1 of 3 answered', got: %s", view)
	}
}

func TestViewShowsCtrlCHint(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:         true,
		ctrlCPending: true,
		viewport:     vp,
		input:        textinput.New(),
		width:        80,
	}
	view := m.View()
	if !strings.Contains(view, "Press Ctrl+C again to quit") {
		t.Errorf("view should show ctrl+c hint when pending, got: %s", view)
	}
}

func TestViewNoCtrlCHintWhenIdle(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:     true,
		viewport: vp,
		input:    textinput.New(),
		width:    80,
	}
	view := m.View()
	if strings.Contains(view, "Press Ctrl+C again to quit") {
		t.Error("view should NOT show ctrl+c hint when not pending")
	}
}

func TestDoneViewContainsInputField(t *testing.T) {
	input := textinput.New()
	input.Placeholder = "Type here..."
	vp := viewport.New(80, 20)
	m := model{
		done:     true,
		viewport: vp,
		input:    input,
		width:    80,
	}
	view := m.View()
	if !strings.Contains(view, "Type here...") {
		t.Error("done view should show the text input field")
	}
}

func TestViewShowsSpinnerWhenRunning(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:      false,
		spinFrame: 0,
		viewport:  vp,
		input:     textinput.New(),
		width:     80,
	}
	view := m.View()
	if !strings.Contains(view, "⠋") {
		t.Errorf("expected spinner char in view when running, got: %s", view)
	}
}

func TestViewShowsSpinnerWithPendingQuestions(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:      false,
		spinFrame: 0,
		viewport:  vp,
		input:     textinput.New(),
		width:     80,
		questions: []pendingQuestion{{ID: "1", Question: "What platform?"}},
	}
	view := m.View()
	if !strings.Contains(view, "⠋") {
		t.Errorf("expected spinner char when questions pending, got: %s", view)
	}
	if !strings.Contains(view, "Pending:") {
		t.Errorf("expected 'Pending:' when questions present, got: %s", view)
	}
}

func TestViewNoSpinnerWhenDone(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:      true,
		spinFrame: 3,
		viewport:  vp,
		input:     textinput.New(),
		width:     80,
	}
	view := m.View()
	for _, c := range SpinnerChars {
		if strings.Contains(view, c) {
			t.Errorf("view should not contain spinner char %q when done", c)
		}
	}
}

func TestFreshSessionViewShowsYouPrefix(t *testing.T) {
	vp := viewport.New(80, 20)
	feature := "a cli that fetch todos from my personal website"
	vp.SetContent(WrapText("[You] "+feature, vp.Width))

	m := model{
		feature:  feature,
		done:     false,
		viewport: vp,
		input:    textinput.New(),
		width:    80,
		logs:     []string{"[You] " + feature},
	}
	view := m.View()
	if !strings.Contains(view, "[You] a cli that fetch todos") {
		t.Errorf("fresh session view should contain '[You] a cli that fetch todos', got: %s", view)
	}
}

func TestResumeSessionLogsStartWithExistingContent(t *testing.T) {
	existingLogs := []string{
		"[You] Add dark mode",
		"💬   ASSISTANT",
		"  Understood, let me brainstorm this feature.",
	}
	vp := viewport.New(80, 20)
	vp.SetContent(WrapText(strings.Join(existingLogs, "\n"), vp.Width))

	m := model{
		feature:  "Add dark mode",
		done:     false,
		viewport: vp,
		input:    textinput.New(),
		width:    80,
		logs:     existingLogs,
	}
	view := m.View()
	count := strings.Count(view, "[You]")
	if count != 1 {
		t.Errorf("resume session should have exactly 1 '[You]', got %d", count)
	}
}

func TestClarificationViewShowsHighlightedOption(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Payment?", Options: []QuestionOption{{Label: "Card"}, {Label: "PayPal"}, {Label: "Bank"}}},
		},
		optionHighlightIdx: 1,
	}
	view := m.View()
	if strings.Count(view, "▸ Card") > 0 {
		t.Error("Card should not show ▸ when index 1 highlighted")
	}
	if !strings.Contains(view, "▸ PayPal") {
		t.Error("PayPal should show ▸ when index 1 highlighted")
	}
	if strings.Count(view, "▸ Bank") > 0 {
		t.Error("Bank should not show ▸ when index 1 highlighted")
	}
}

func TestClarificationViewDimsEmptyInput(t *testing.T) {
	vp := viewport.New(80, 20)
	input := textinput.New()
	input.Placeholder = "default"
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             input,
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}}},
		},
	}
	view := m.View()
	if !strings.Contains(view, "Select an option") {
		t.Error("should show dimmed input placeholder when on real option")
	}
	if !strings.Contains(view, "Shift+Tab") {
		t.Error("should show Shift+Tab in status line")
	}
}

func TestClarificationViewNormalInputWithText(t *testing.T) {
	vp := viewport.New(80, 20)
	input := textinput.New()
	input.SetValue("My answer")
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             input,
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?"},
		},
	}
	view := m.View()
	if !strings.Contains(view, "My answer") {
		t.Errorf("view should contain typed text 'My answer', got: %s", view)
	}
}

func TestClarificationViewShowsVirtualOption(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}, {Label: "B"}}},
		},
		optionHighlightIdx: 0,
	}
	view := m.View()
	if !strings.Contains(view, "Type your answer below") {
		t.Error("view should show virtual 'Type your answer below' option")
	}
}

func TestClarificationVirtualOptionHighlighted(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}}},
		},
		optionHighlightIdx: 1,
	}
	view := m.View()
	if !strings.Contains(view, "▸ ✎") {
		t.Error("virtual option should show ▸ when highlighted")
	}
}

func TestClarificationViewDimsInputOnRealOption(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}}},
		},
		optionHighlightIdx: 0,
	}
	view := m.View()
	if !strings.Contains(view, "Select an option") {
		t.Error("should show dimmed placeholder when on real option")
	}
}

func TestClarificationViewActiveInputOnVirtualOption(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		width:             80,
		questions: []pendingQuestion{
			{ID: "1", Question: "Q?", Options: []QuestionOption{{Label: "A"}}},
		},
		optionHighlightIdx: 1,
	}
	view := m.View()
	if !strings.Contains(view, "Type your answer") {
		t.Error("should show active placeholder when on virtual option")
	}
}
