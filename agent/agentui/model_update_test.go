package agentui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDoneStateDoubleCtrlCQuits(t *testing.T) {
	m := model{done: true}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	um := updated.(model)
	if !um.ctrlCPending {
		t.Error("first ctrl+c should set ctrlCPending=true")
	}
	if cmd == nil {
		t.Error("first ctrl+c should return a non-nil timer command")
	}

	updated, cmd = um.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("second ctrl+c should return a command")
		return
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("second ctrl+c should return tea.Quit")
	}
}

func TestDoneStateCtrlCResetOnOtherKey(t *testing.T) {
	m := model{done: true, ctrlCPending: true}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	um := updated.(model)
	if um.ctrlCPending {
		t.Error("any other key should reset ctrlCPending")
	}
}

func TestClarificationDoubleCtrlCQuits(t *testing.T) {
	m := model{
		done:              true,
		clarificationMode: true,
		questions:         []pendingQuestion{{ID: "1", Question: "Q1"}},
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	um := updated.(model)
	if !um.ctrlCPending {
		t.Error("first ctrl+c should set ctrlCPending=true")
	}
	if cmd == nil {
		t.Error("first ctrl+c should return timer command")
	}

	updated, cmd = um.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("second ctrl+c should return a command")
		return
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("second ctrl+c should return tea.Quit")
	}
}

func TestCtrlCResetOnTimeout(t *testing.T) {
	m := model{done: true, ctrlCPending: true}
	updated, _ := m.Update(ctrlCResetMsg{})
	um := updated.(model)
	if um.ctrlCPending {
		t.Error("ctrlCResetMsg should set ctrlCPending=false")
	}
}

func TestRunningDoubleCtrlCQuits(t *testing.T) {
	vp := viewport.New(80, 20)
	m := model{
		done:     false,
		viewport: vp,
		input:    textinput.New(),
		width:    80,
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	um := updated.(model)
	if !um.ctrlCPending {
		t.Error("first ctrl+c while running should set ctrlCPending=true")
	}
	if cmd == nil {
		t.Error("first ctrl+c should return timer command")
	}

	updated, cmd = um.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("second ctrl+c should quit")
		return
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("second ctrl+c should return tea.Quit")
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

	testMsg := llmDoneMsg{Output: "test output", Err: nil}
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
	if msg.Output != "test output" {
		t.Errorf("expected output 'test output', got %q", msg.Output)
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

	if cmd == nil {
		t.Error("expected a waitForLLMDone command for follow-up")
	}
}

func TestSpinnerCharWrapsAround(t *testing.T) {
	if SpinnerChar(0) != "⠋" {
		t.Errorf("expected first char '⠋', got %q", SpinnerChar(0))
	}
	if SpinnerChar(len(SpinnerChars)) != "⠋" {
		t.Errorf("expected wrap-around to '⠋', got %q", SpinnerChar(len(SpinnerChars)))
	}
	if SpinnerChar(len(SpinnerChars)-1) != "⠏" {
		t.Errorf("expected last char '⠏', got %q", SpinnerChar(len(SpinnerChars)-1))
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

func TestExitCommandWhenDone(t *testing.T) {
	input := textinput.New()
	input.SetValue("/exit")
	m := model{done: true, input: input}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("/exit should return a command")
		return
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("/exit when done should return tea.Quit")
	}
}

func TestExitCommandWhenNotDone(t *testing.T) {
	input := textinput.New()
	input.SetValue("/exit")
	m := model{done: false, input: input}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("/exit should return a command")
		return
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("/exit when not done should return tea.Quit")
	}
}

func TestExitCommandWithWhitespace(t *testing.T) {
	input := textinput.New()
	input.SetValue("  /exit  ")
	m := model{done: true, input: input}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("/exit with whitespace should return a command")
		return
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("/exit with whitespace should return tea.Quit")
	}
}

func TestExitCommandNotTriggeredMidText(t *testing.T) {
	input := textinput.New()
	input.SetValue("some /exit thing")
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
		t.Error("'some /exit thing' should trigger follow-up, not quit — expected done=false")
	}
}
