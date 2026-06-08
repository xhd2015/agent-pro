package agentui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/xhd2015/agent-pro/agent/session"
)

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
			got := WrapText(tt.input, tt.width)
			if got != tt.expect {
				t.Errorf("WrapText(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expect)
			}
		})
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

func TestHasUnansweredQuestions(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
		},
	}
	if !m.hasUnansweredQuestions() {
		t.Error("expected hasUnansweredQuestions=true when Q2 has no answer")
	}

	m.questions[1].Answer = "A2"
	if m.hasUnansweredQuestions() {
		t.Error("expected hasUnansweredQuestions=false when all answered")
	}
}

func TestAnsweredCount(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
			{ID: "3", Question: "Q3", Answer: "A3"},
		},
	}
	if n := m.answeredCount(); n != 2 {
		t.Errorf("expected answeredCount=2, got %d", n)
	}
}

func TestCurrentPendingQuestionSkipsAnswered(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
			{ID: "3", Question: "Q3", Answer: "A3"},
		},
		selIdx: 0,
	}
	q := m.currentPendingQuestion()
	if q == nil {
		t.Fatal("expected a pending question")
	}
	if q.ID != "2" {
		t.Errorf("expected Q2 (unanswered), got %s", q.ID)
	}
}

func TestCurrentPendingQuestionAllAnswered(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
		},
		selIdx: 0,
	}
	q := m.currentPendingQuestion()
	if q != nil {
		t.Errorf("expected nil when all answered, got %v", q)
	}
}

func TestCurrentPendingQuestionEmpty(t *testing.T) {
	m := &model{selIdx: 0}
	q := m.currentPendingQuestion()
	if q != nil {
		t.Errorf("expected nil for empty questions, got %v", q)
	}
}

func TestSubmitAnswerRecordsAnswer(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
		},
		selIdx: 0,
		logs:   []string{},
	}
	m.submitAnswer("my answer")
	if m.questions[0].Answer != "my answer" {
		t.Errorf("expected answer 'my answer', got %q", m.questions[0].Answer)
	}
	found := false
	for _, log := range m.logs {
		if strings.Contains(log, "[You] Answered #1: my answer") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected answer log entry")
	}
}

func TestSubmitAnswerClearsInput(t *testing.T) {
	input := textinput.New()
	input.SetValue("typed answer")
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
		},
		selIdx: 0,
		logs:   []string{},
		input:  input,
	}
	m.submitAnswer("typed answer")
	if m.input.Value() != "" {
		t.Errorf("expected input cleared after submit, got %q", m.input.Value())
	}
}

func TestSubmitAnswerAllAnsweredTransitionsOutOfClarification(t *testing.T) {
	logCh := make(chan string, 64)
	doneCh := make(chan llmDoneMsg, 1)
	vp := viewport.New(80, 20)
	m := &model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
		},
		selIdx:     1,
		llmModel:   "test-model",
		sessionID:  "test-session",
		sessionDir: t.TempDir(),
		logCh:      logCh,
		llmDoneCh:  doneCh,
		width:      80,
		logs:       []string{},
	}
	m.submitAnswer("A2")
	if m.clarificationMode {
		t.Error("expected clarificationMode=false after all answered")
	}
	if m.done {
		t.Error("expected done=false after all answered (resumes LLM)")
	}
}

func TestSubmitAnswerStillHasUnanswered(t *testing.T) {
	vp := viewport.New(80, 20)
	m := &model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
			{ID: "2", Question: "Q2"},
		},
		selIdx: 0,
		width:  80,
		logs:   []string{},
	}
	m.submitAnswer("A1")
	if !m.clarificationMode {
		t.Error("expected clarificationMode to stay true when more unanswered")
	}
	if m.selIdx != 1 {
		t.Errorf("expected selIdx=1 after answering Q1, got %d", m.selIdx)
	}
}

func TestBuildResumePrompt(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Payment methods?", Options: []QuestionOption{{Label: "Card"}, {Label: "PayPal"}}, Answer: "Card"},
			{ID: "2", Question: "Max amount?", Answer: "5000"},
			{ID: "3", Question: "Unanswered?", Answer: ""},
		},
	}
	prompt := m.buildResumePrompt("follow up text")
	if !strings.Contains(prompt, "Payment methods?") {
		t.Error("prompt should contain answered question")
	}
	if !strings.Contains(prompt, "Card") {
		t.Error("prompt should contain answer 'Card'")
	}
	if !strings.Contains(prompt, "5000") {
		t.Error("prompt should contain answer '5000'")
	}
	if !strings.Contains(prompt, "Options: Card, PayPal") {
		t.Error("prompt should list options")
	}
	if strings.Contains(prompt, "Unanswered?") {
		t.Error("prompt should NOT contain unanswered question")
	}
	if !strings.Contains(prompt, "follow up text") {
		t.Error("prompt should contain follow-up text")
	}
	if !strings.Contains(prompt, "## Follow-up") {
		t.Error("prompt should have ## Follow-up section when followUp is provided")
	}
}

func TestBuildResumePromptNoFollowUp(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1?", Answer: "A1"},
		},
	}
	prompt := m.buildResumePrompt("")
	if strings.Contains(prompt, "## Follow-up") {
		t.Error("prompt should NOT have ## Follow-up when followUp is empty")
	}
}

func TestLlmdoneSetsClarificationMode(t *testing.T) {
	logCh := make(chan string, 64)
	vp := viewport.New(80, 20)
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
		},
		viewport: vp,
		logCh:    logCh,
		logs:     []string{},
		width:    80,
	}
	updated, _ := m.Update(llmDoneMsg{output: "test output"})
	um := updated.(model)
	if !um.clarificationMode {
		t.Error("expected clarificationMode=true when pending questions exist")
	}
	if !um.done {
		t.Error("expected done=true after llmDoneMsg")
	}
}

func TestLlmdoneNoQuestionsGoesToDone(t *testing.T) {
	logCh := make(chan string, 64)
	vp := viewport.New(80, 20)
	m := &model{
		viewport: vp,
		logCh:    logCh,
		logs:     []string{},
		width:    80,
	}
	updated, _ := m.Update(llmDoneMsg{output: "test output"})
	um := updated.(model)
	if um.clarificationMode {
		t.Error("expected clarificationMode=false when no pending questions")
	}
	if !um.done {
		t.Error("expected done=true after llmDoneMsg")
	}
}

func TestLlmdoneAllAnsweredGoesToDone(t *testing.T) {
	logCh := make(chan string, 64)
	vp := viewport.New(80, 20)
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
		},
		viewport: vp,
		logCh:    logCh,
		logs:     []string{},
		width:    80,
	}
	updated, _ := m.Update(llmDoneMsg{output: "test output"})
	um := updated.(model)
	if um.clarificationMode {
		t.Error("expected clarificationMode=false when all answered")
	}
	if !um.done {
		t.Error("expected done=true after llmDoneMsg")
	}
}

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

func TestNewSessionIDFormat(t *testing.T) {
	id := newSessionID("tcd_")
	if !strings.HasPrefix(id, "tcd_") {
		t.Errorf("session ID should start with 'tcd_', got %q", id)
	}
	if len(id) < 20 {
		t.Errorf("session ID too short: %s", id)
	}
	id2 := newSessionID("tcd_")
	if id == id2 {
		t.Error("newSessionID should produce unique IDs")
	}
}

func TestReadSessionFromDir(t *testing.T) {
	dir := t.TempDir()

	meta := sessionMeta{SessionID: "sid_1", Feature: "Test feature", Model: "gpt-4o"}
	session.WriteJSON(dir, "metadata.json", meta)
	session.AppendLine(dir, "events.jsonl", `{"type":"text","timestamp":1,"sessionID":"sid_1","part":{"text":"hello"}}`)
	session.AppendLine(dir, "events.jsonl", `{"type":"tool_use","timestamp":2,"sessionID":"sid_1","part":{"tool":"bash","state":{"status":"completed","title":"ls"}}}`)

	sid, _, feat, model, logs := readSessionFromDir(dir, "sid_1")
	if sid != "sid_1" {
		t.Errorf("expected session ID 'sid_1', got %q", sid)
	}
	if feat != "Test feature" {
		t.Errorf("expected feature 'Test feature', got %q", feat)
	}
	if model != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got %q", model)
	}
	if len(logs) < 1 {
		t.Error("expected at least one formatted log entry")
	}
}

func TestReadSessionFromDirEmptyEvents(t *testing.T) {
	dir := t.TempDir()
	meta := sessionMeta{SessionID: "sid_2", Feature: "X", Model: "Y"}
	session.WriteJSON(dir, "metadata.json", meta)

	sid, _, feat, model, logs := readSessionFromDir(dir, "sid_2")
	if sid != "sid_2" || feat != "X" || model != "Y" {
		t.Errorf("unexpected values: sid=%s feat=%s model=%s", sid, feat, model)
	}
	if logs != nil {
		t.Error("expected nil logs for empty events file")
	}
}

func TestReadSessionFromDirNonExistent(t *testing.T) {
	sid, _, feat, model, logs := readSessionFromDir("/nonexistent/path", "")
	if sid != "" || feat != "" || model != "" || logs != nil {
		t.Error("expected empty result for non-existent dir")
	}
}

func TestReadSessionFromDirFallbackSessionID(t *testing.T) {
	dir := t.TempDir()
	meta := sessionMeta{SessionID: "", Feature: "F", Model: "M"}
	session.WriteJSON(dir, "metadata.json", meta)

	sid, _, _, _, _ := readSessionFromDir(dir, "fallback-id")
	if sid != "fallback-id" {
		t.Errorf("expected fallback session ID 'fallback-id', got %q", sid)
	}
}

func TestResolveSessionResumes(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	resumeID := "tcd_abc123"
	dir, err := session.Dir("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	session.WriteJSON(dir, "metadata.json", sessionMeta{
		SessionID: resumeID,
		Feature:   "Add dark mode",
		Model:     "gpt-4o",
	})
	session.AppendLine(dir, "events.jsonl", `{"type":"text","timestamp":1,"sessionID":"tcd_abc123","part":{"text":"hello"}}`)

	sid, _, sdir, feat, model, logs, err := resolveSession("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("resolveSession: %v", err)
	}
	if sid != resumeID {
		t.Errorf("expected session ID %q, got %q", resumeID, sid)
	}
	if sdir != dir {
		t.Errorf("expected dir %q, got %q", dir, sdir)
	}
	if feat != "Add dark mode" {
		t.Errorf("expected feature 'Add dark mode', got %q", feat)
	}
	if model != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got %q", model)
	}
	if len(logs) == 0 {
		t.Error("expected at least one log entry")
	}
}

func TestResolveSessionNotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	_, _, _, _, _, _, err := resolveSession("test-agent", "nonexistent_id")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestResolveSessionEmptyID(t *testing.T) {
	sid, _, _, feat, model, _, err := resolveSession("test-agent", "")
	if err != nil {
		t.Fatalf("resolveSession empty: %v", err)
	}
	if sid != "" {
		t.Errorf("expected empty session ID, got %q", sid)
	}
	if feat != "" || model != "" {
		t.Error("expected empty feature and model for empty resume ID")
	}
}

func TestResolveSessionModelOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	resumeID := "tcd_xyz"
	dir, err := session.Dir("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	session.WriteJSON(dir, "metadata.json", sessionMeta{
		SessionID: resumeID,
		Feature:   "Feature X",
		Model:     "claude-3",
	})

	_, _, _, _, model, _, err := resolveSession("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("resolveSession: %v", err)
	}
	if model != "claude-3" {
		t.Errorf("expected model from metadata 'claude-3', got %q", model)
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

func TestConfigRunHonorsAgentName(t *testing.T) {
	cfg := Config{
		AgentName:     "my-agent",
		SessionPrefix: "ma_",
		Prompt:        "prompt",
		Usage:         "my usage",
	}
	if cfg.AgentName != "my-agent" {
		t.Errorf("expected AgentName 'my-agent', got %q", cfg.AgentName)
	}
	if cfg.SessionPrefix != "ma_" {
		t.Errorf("expected SessionPrefix 'ma_', got %q", cfg.SessionPrefix)
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

func TestLoadQuestionsFromFile(t *testing.T) {
	dir := t.TempDir()
	session.AppendLine(dir, "questions.jsonl", `{"type":"question","id":"1","question":"Q1","options":[{"label":"A"},{"label":"B"}]}`)
	session.AppendLine(dir, "questions.jsonl", `{"type":"question","id":"2","question":"Q2"}`)
	session.AppendLine(dir, "questions.jsonl", `{"type":"answer","id":"1","answer":"Answer 1"}`)

	m := &model{
		sessionDir: dir,
		logs:       []string{},
	}
	m.loadQuestionsFromFile()

	if len(m.questions) != 2 {
		t.Fatalf("expected 2 questions loaded, got %d", len(m.questions))
	}
	if m.questions[0].ID != "1" || m.questions[0].Question != "Q1" {
		t.Errorf("Q1 not loaded correctly: %+v", m.questions[0])
	}
	if m.questions[0].Answer != "Answer 1" {
		t.Errorf("expected Q1 answer 'Answer 1', got %q", m.questions[0].Answer)
	}
	if m.questions[1].Answer != "" {
		t.Errorf("expected Q2 answer empty, got %q", m.questions[1].Answer)
	}
	if len(m.questions[0].Options) != 2 {
		t.Errorf("expected 2 options for Q1, got %d", len(m.questions[0].Options))
	}
}

func TestLoadQuestionsFromFileNoFile(t *testing.T) {
	m := &model{
		sessionDir: "/nonexistent",
		logs:       []string{},
	}
	m.loadQuestionsFromFile()
	if len(m.questions) != 0 {
		t.Errorf("expected 0 questions for missing file, got %d", len(m.questions))
	}
}

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
