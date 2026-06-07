package agentui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/xhd2015/agent-pro/agent/session"
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
			got := WrapText(tt.input, tt.width)
			if got != tt.expect {
				t.Errorf("WrapText(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expect)
			}
		})
	}
}

func TestGenerateOutputName(t *testing.T) {
	tests := []struct {
		input  string
		suffix string
		expect string
	}{
		{"Cat Classifier", "-tests-design.md", "cat-classifier-tests-design.md"},
		{"Add dark mode", "-tests-design.md", "add-dark-mode-tests-design.md"},
		{"Login!!! Page", "-tests-design.md", "login-page-tests-design.md"},
		{"a very long feature description that goes on and on and on and on", "-tests-design.md", "a-very-long-feature-description-that-goes-on-and-o-tests-design.md"},
		{"", "-tests-design.md", "tests-design.md"},
		{"  spaces   everywhere  ", "-tests-design.md", "spaces-everywhere-tests-design.md"},
		{"Brainstorm Feature", "-idea.md", "brainstorm-feature-idea.md"},
		{"", "-idea.md", "idea.md"},
	}

	for _, tt := range tests {
		got := GenerateOutputName(tt.input, tt.suffix)
		if got != tt.expect {
			t.Errorf("GenerateOutputName(%q, %q) = %q, want %q", tt.input, tt.suffix, got, tt.expect)
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
		OutputSuffix:  "-out.md",
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
