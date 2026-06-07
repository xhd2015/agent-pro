package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestParseSessionFile(t *testing.T) {
	data := []byte(`
{"session_id":"abc123","feature":"Add dark mode","model":"gpt-4o"}
{"type":"text","timestamp":1,"sessionID":"abc123","part":{"text":"Hello"}}
{"type":"tool_use","timestamp":2,"sessionID":"abc123","part":{"tool":"bash","state":{"status":"completed","title":"ls -la"}}}
`)

	sessionID, feature, model, logs, err := parseSessionFile(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessionID != "abc123" {
		t.Errorf("expected sessionID abc123, got %s", sessionID)
	}
	if feature != "Add dark mode" {
		t.Errorf("expected feature 'Add dark mode', got %s", feature)
	}
	if model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", model)
	}
	if len(logs) == 0 {
		t.Error("expected at least one formatted log entry")
	}
	if !strings.Contains(logs[0], "Hello") {
		t.Errorf("expected first log to contain 'Hello', got: %s", logs[0])
	}
}

func TestParseSessionFileEmpty(t *testing.T) {
	_, _, _, _, err := parseSessionFile([]byte(""))
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestParseSessionFileEmptySessionID(t *testing.T) {
	// metadata has empty session_id, sessionID comes from event lines
	data := []byte(`
{"session_id":"","feature":"Add dark mode","model":"gpt-4o"}
{"type":"text","timestamp":1,"sessionID":"ses_abc123","part":{"text":"Hello"}}
`)
	sessionID, feature, model, logs, err := parseSessionFile(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessionID != "ses_abc123" {
		t.Errorf("expected sessionID ses_abc123, got %s", sessionID)
	}
	if feature != "Add dark mode" {
		t.Errorf("expected feature 'Add dark mode', got %s", feature)
	}
	if model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", model)
	}
	if len(logs) == 0 || !strings.Contains(logs[0], "Hello") {
		t.Error("expected logs with Hello")
	}
}

func TestParseSessionFileNoSessionID(t *testing.T) {
	// no session_id in metadata, no sessionID in events
	data := []byte(`
{"session_id":"","feature":"X","model":"Y"}
`)
	_, _, _, _, err := parseSessionFile(data)
	if err == nil {
		t.Error("expected error when no sessionID found")
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		width  int
		expect string
	}{
		{
			name:   "no wrap needed",
			input:  "hello world",
			width:  20,
			expect: "hello world",
		},
		{
			name:   "wrap at word boundary",
			input:  "hello world here",
			width:  11,
			expect: "hello world\nhere",
		},
		{
			name:   "multiline input",
			input:  "line one\nline two long",
			width:  8,
			expect: "line one\nline two\nlong",
		},
		{
			name:   "long word forced break",
			input:  "supercalifragilistic",
			width:  10,
			expect: "supercalif\nragilistic",
		},
		{
			name:   "zero width passes through",
			input:  "hello world",
			width:  0,
			expect: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.input, tt.width)
			if got != tt.expect {
				t.Errorf("wrapText(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expect)
			}
		})
	}
}

func TestGenerateOutputName(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"Cat Classifier", "cat-classifier-tests-design.md"},
		{"Add dark mode", "add-dark-mode-tests-design.md"},
		{"Login!!! Page", "login-page-tests-design.md"},
		{"a very long feature description that goes on and on and on and on", "a-very-long-feature-description-that-goes-on-and-o-tests-design.md"},
		{"", "tests-design.md"},
		{"  spaces   everywhere  ", "spaces-everywhere-tests-design.md"},
	}

	for _, tt := range tests {
		got := generateOutputName(tt.input)
		if got != tt.expect {
			t.Errorf("generateOutputName(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}

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

func TestNonDoneViewShowsPendingQuestions(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:     false,
		viewport: vp,
		input:    textinput.New(),
		width:    80,
	}
	view := m.View()
	if !strings.Contains(view, "Waiting for agent questions") {
		t.Errorf("non-done view should show 'Waiting for agent questions', got: %s", view)
	}
}

func TestDoneStateCtrlCQuits(t *testing.T) {
	m := model{done: true}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c when done should return a non-nil command")
		return
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Error("ctrl+c when done should return tea.Quit")
	}
}

func TestDoneStateEnterEmptyStaysDone(t *testing.T) {
	input := textinput.New()
	input.SetValue("")
	m := model{
		done:  true,
		input: input,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !updated.(model).done {
		t.Error("enter with empty input when done should keep done=true")
	}
}

func TestDoneStateEnterWithTextResumes(t *testing.T) {
	input := textinput.New()
	input.SetValue("Can you add more tests?")
	logCh := make(chan string, 64)
	doneCh := make(chan llmDoneMsg, 1)

	vp := viewport.New(80, 20)
	m := model{
		done:      true,
		input:     input,
		llmModel:  "test-model",
		sessionID: "test-session",
		logCh:     logCh,
		llmDoneCh: doneCh,
		viewport:  vp,
		width:     80,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	um := updated.(model)
	if um.done {
		t.Error("enter with follow-up text should set done=false")
	}

	found := false
	for _, log := range um.logs {
		if strings.Contains(log, "[You] Can you add more tests?") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected follow-up text '[You] Can you add more tests?' in logs")
	}
}

func TestDoneStateKeyInputPassesToTextinput(t *testing.T) {
	input := textinput.New()
	m := model{
		done:  true,
		input: input,
	}
	// Sending a backspace key (non-enter, non-ctrl+c) should be routed
	// to the textinput and should NOT change done state.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	um := updated.(model)
	if !um.done {
		t.Error("typing when done should keep done=true")
	}
	if um.input.Value() != "" {
		t.Error("backspace on empty textinput should not crash")
	}
}

func TestWaitForLLMDoneReadsFromChannel(t *testing.T) {
	doneCh := make(chan llmDoneMsg, 1)
	m := model{llmDoneCh: doneCh}

	testMsg := llmDoneMsg{output: "test output", err: nil}
	doneCh <- testMsg

	cmd := m.waitForLLMDone()
	if cmd == nil {
		t.Fatal("waitForLLMDone should return a non-nil command")
	}
	result := cmd()
	msg, ok := result.(llmDoneMsg)
	if !ok {
		t.Fatalf("expected llmDoneMsg, got %T", result)
	}
	if msg.output != "test output" {
		t.Errorf("expected output 'test output', got %q", msg.output)
	}
}

func TestDoneStateEnterSetsNewDoneChannel(t *testing.T) {
	input := textinput.New()
	input.SetValue("follow-up message")
	logCh := make(chan string, 64)
	oldDoneCh := make(chan llmDoneMsg, 1)

	vp := viewport.New(80, 20)
	m := model{
		done:      true,
		input:     input,
		llmModel:  "test-model",
		sessionID: "test-session",
		logCh:     logCh,
		llmDoneCh: oldDoneCh,
		viewport:  vp,
		width:     80,
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(model)

	if um.llmDoneCh == oldDoneCh {
		t.Error("expected a new llmDoneCh to be created for follow-up")
	}

	// The returned command should read from the new channel
	if cmd == nil {
		t.Error("expected a waitForLLMDone command for follow-up")
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

func TestSpinnerCharWrapsAround(t *testing.T) {
	if spinnerChar(0) != "⠋" {
		t.Errorf("expected first char '⠋', got %q", spinnerChar(0))
	}
	if spinnerChar(len(spinnerChars)) != "⠋" {
		t.Errorf("expected wrap-around to '⠋', got %q", spinnerChar(len(spinnerChars)))
	}
	if spinnerChar(len(spinnerChars)-1) != "⠏" {
		t.Errorf("expected last char '⠏', got %q", spinnerChar(len(spinnerChars)-1))
	}
}

func TestTickIncrementsSpinFrame(t *testing.T) {
	m := model{done: false}
	updated, _ := m.Update(tickMsg{})
	um := updated.(model)
	if um.spinFrame != 1 {
		t.Errorf("expected spinFrame=1 after tick, got %d", um.spinFrame)
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
		questions: []pendingQuestion{{id: "1", question: "What platform?"}},
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
	for _, c := range spinnerChars {
		if strings.Contains(view, c) {
			t.Errorf("view should not contain spinner char %q when done", c)
		}
	}
}

func TestFollowUpResumesSpinner(t *testing.T) {
	input := textinput.New()
	input.SetValue("continue please")
	logCh := make(chan string, 64)
	doneCh := make(chan llmDoneMsg, 1)
	vp := viewport.New(80, 20)

	m := model{
		done:      true,
		input:     input,
		llmModel:  "test-model",
		sessionID: "test-session",
		logCh:     logCh,
		llmDoneCh: doneCh,
		viewport:  vp,
		spinFrame: 5,
		width:     80,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(model)
	if um.done {
		t.Error("follow-up should set done=false")
	}
	if um.spinFrame != 5 {
		t.Errorf("spinFrame should be preserved after follow-up, got %d", um.spinFrame)
	}
}
