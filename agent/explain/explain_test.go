package explain

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type mockRunner struct {
	startCalled   int
	resumeCalled  int
	lastModel     string
	lastPrompt    string
	lastMeta      json.RawMessage
	startOutputs  []mockRunOutput
	resumeOutputs []mockRunOutput
}

type mockRunOutput struct {
	sessionID string
	output    string
	err       error
}

func (m *mockRunner) Start(ctx context.Context, model string, prompt string) (string, string, error) {
	m.startCalled++
	m.lastModel = model
	m.lastPrompt = prompt
	if len(m.startOutputs) == 0 {
		return "mock-session-id", "mock output for: " + prompt, nil
	}
	idx := m.startCalled - 1
	if idx >= len(m.startOutputs) {
		idx = len(m.startOutputs) - 1
	}
	out := m.startOutputs[idx]
	return out.sessionID, out.output, out.err
}

func (m *mockRunner) Resume(ctx context.Context, model string, prompt string, meta json.RawMessage) (string, error) {
	m.resumeCalled++
	m.lastModel = model
	m.lastPrompt = prompt
	m.lastMeta = meta
	if len(m.resumeOutputs) == 0 {
		return "mock resumed: " + prompt, nil
	}
	idx := m.resumeCalled - 1
	if idx >= len(m.resumeOutputs) {
		idx = len(m.resumeOutputs) - 1
	}
	out := m.resumeOutputs[idx]
	return out.output, out.err
}

func TestRunExplain_NewSession_SingleArg(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-new-1", output: "run的过去式是ran"},
		},
	}

	err := RunExplainWithRunner([]string{"run的过去式"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 1 {
		t.Fatalf("expected 1 start call, got %d", runner.startCalled)
	}
	if runner.resumeCalled != 0 {
		t.Fatalf("expected 0 resume calls, got %d", runner.resumeCalled)
	}
	if runner.lastPrompt != "run的过去式" {
		t.Fatalf("expected prompt 'run的过去式', got %q", runner.lastPrompt)
	}
}

func TestRunExplain_ResumeSession_StrictPrefixMatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-run", SessionData{
		AgentRunner: "opencode",
		Model:       "deepseek",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess-existing"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run"},
			{Role: "assistant", Message: "ran"},
		},
	})

	runner := &mockRunner{
		resumeOutputs: []mockRunOutput{
			{output: "ran是过去式"},
		},
	}

	err := RunExplainWithRunner([]string{"run的过去式", "ran是什么意思"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 0 {
		t.Fatalf("expected 0 start calls, got %d", runner.startCalled)
	}
	if runner.resumeCalled != 1 {
		t.Fatalf("expected 1 resume call, got %d", runner.resumeCalled)
	}
	if runner.lastPrompt != "ran是什么意思" {
		t.Fatalf("expected prompt 'ran是什么意思', got %q", runner.lastPrompt)
	}

	var opencodeMeta struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(runner.lastMeta, &opencodeMeta); err != nil {
		t.Fatalf("unmarshal resume meta: %v", err)
	}
	if opencodeMeta.SessionID != "sess-existing" {
		t.Fatalf("expected session_id sess-existing, got %s", opencodeMeta.SessionID)
	}
}

func TestRunExplain_PrefixMatchResume(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-run-de-guo-qu", SessionData{
		AgentRunner: "opencode",
		Model:       "deepseek",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess-existing"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的过去式"},
			{Role: "assistant", Message: "ran"},
		},
	})

	runner := &mockRunner{
		resumeOutputs: []mockRunOutput{
			{output: "ran, running, runs"},
		},
	}

	err := RunExplainWithRunner([]string{"run的过去式和各种形态"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 0 {
		t.Fatalf("expected 0 start calls (should resume via prefix match), got %d", runner.startCalled)
	}
	if runner.resumeCalled != 1 {
		t.Fatalf("expected 1 resume call, got %d", runner.resumeCalled)
	}
	if runner.lastPrompt != "run的过去式和各种形态" {
		t.Fatalf("expected prompt to be the user input, got %q", runner.lastPrompt)
	}
}

func TestRunExplain_ExactMatchStartsNewSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-exact", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess-exact"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的过去式"},
			{Role: "assistant", Message: "ran"},
		},
	})

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-new-2", output: "ran is past tense"},
		},
	}

	err := RunExplainWithRunner([]string{"run的过去式"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 1 {
		t.Fatalf("expected 1 start call (exact match = new session), got %d", runner.startCalled)
	}
	if runner.resumeCalled != 0 {
		t.Fatalf("expected 0 resume calls, got %d", runner.resumeCalled)
	}
}

func TestRunExplain_NoMatchNewSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-unrelated", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess-old"}),
		},
		Messages: []Message{
			{Role: "user", Message: "python的用法"},
			{Role: "assistant", Message: "python is..."},
		},
	})

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-new-2", output: "ran是run的过去式"},
		},
	}

	err := RunExplainWithRunner([]string{"run的过去式"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 1 {
		t.Fatalf("expected 1 start call, got %d", runner.startCalled)
	}
	if runner.resumeCalled != 0 {
		t.Fatalf("expected 0 resume calls, got %d", runner.resumeCalled)
	}
}

func TestRunExplain_FollowUpOnNewSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-new", output: "run的过去式是ran"},
		},
		resumeOutputs: []mockRunOutput{
			{output: "ran是过去式"},
		},
	}

	err := RunExplainWithRunner([]string{"run的过去式", "ran是什么意思"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 1 {
		t.Fatalf("expected 1 start call, got %d", runner.startCalled)
	}
	if runner.resumeCalled != 1 {
		t.Fatalf("expected 1 resume call (for follow-up), got %d", runner.resumeCalled)
	}
	if runner.lastPrompt != "ran是什么意思" {
		t.Fatalf("expected resume prompt 'ran是什么意思', got %q", runner.lastPrompt)
	}
}

func TestRunExplain_ModelFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-model-test", output: "answer"},
		},
	}

	err := RunExplainWithRunner([]string{"--model", "gpt-4", "hello"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.lastModel != "gpt-4" {
		t.Fatalf("expected model gpt-4, got %q", runner.lastModel)
	}
}

func TestRunExplain_MissingArg(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	runner := &mockRunner{}
	err := RunExplainWithRunner([]string{}, runner)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing message argument") {
		t.Fatalf("expected 'missing message argument', got %q", err.Error())
	}
}

func TestRunExplain_InvalidRunner(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	runner := &mockRunner{}
	err := RunExplainWithRunner([]string{"--agent-runner", "unsupported", "hello"}, runner)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported agent runner") {
		t.Fatalf("expected 'unsupported agent runner', got %q", err.Error())
	}
}
