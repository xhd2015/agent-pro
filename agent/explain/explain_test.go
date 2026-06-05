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

func TestRunExplain_SingleArgNewSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-1", output: "answer"},
		},
	}

	err := RunExplainWithRunner([]string{"hello"}, runner)
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

func TestRunExplain_SingleArgRepeatNewSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-existing", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess-old"}),
		},
		Messages: []Message{
			{Role: "user", Message: "hello"},
			{Role: "assistant", Message: "old answer"},
		},
	})

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-new", output: "new answer"},
		},
	}

	err := RunExplainWithRunner([]string{"hello"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 1 {
		t.Fatalf("expected 1 start call (1 arg never matches), got %d", runner.startCalled)
	}
	if runner.resumeCalled != 0 {
		t.Fatalf("expected 0 resume calls, got %d", runner.resumeCalled)
	}
}

func TestRunExplain_TwoArgsExactMatchResume(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-existing", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess-existing"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A F"},
			{Role: "assistant", Message: "old answer"},
		},
	})

	runner := &mockRunner{
		resumeOutputs: []mockRunOutput{
			{output: "resumed answer"},
		},
	}

	err := RunExplainWithRunner([]string{"A F", "B"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 0 {
		t.Fatalf("expected 0 start calls, got %d", runner.startCalled)
	}
	if runner.resumeCalled != 1 {
		t.Fatalf("expected 1 resume call, got %d", runner.resumeCalled)
	}
	if runner.lastPrompt != "B" {
		t.Fatalf("expected prompt 'B', got %q", runner.lastPrompt)
	}
}

func TestRunExplain_TwoArgsElementMismatchNewSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-existing", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess-old"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A"},
			{Role: "assistant", Message: "ans"},
		},
	})

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-new", output: "new answer"},
		},
		resumeOutputs: []mockRunOutput{
			{output: "follow up answer"},
		},
	}

	err := RunExplainWithRunner([]string{"A F", "B"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 1 {
		t.Fatalf("expected 1 start call, got %d", runner.startCalled)
	}
	if runner.resumeCalled != 1 {
		t.Fatalf("expected 1 resume for follow-up, got %d", runner.resumeCalled)
	}
}

func TestRunExplain_ThreeArgsTwoPrefixResume(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-existing", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess-existing"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A"},
			{Role: "assistant", Message: "ans1"},
			{Role: "user", Message: "B"},
			{Role: "assistant", Message: "ans2"},
		},
	})

	runner := &mockRunner{
		resumeOutputs: []mockRunOutput{
			{output: "resumed C"},
		},
	}

	err := RunExplainWithRunner([]string{"A", "B", "C"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 0 {
		t.Fatalf("expected 0 start calls, got %d", runner.startCalled)
	}
	if runner.resumeCalled != 1 {
		t.Fatalf("expected 1 resume call, got %d", runner.resumeCalled)
	}
	if runner.lastPrompt != "C" {
		t.Fatalf("expected prompt 'C', got %q", runner.lastPrompt)
	}
}

func TestRunExplain_NoMatchNewSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-existing", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess-old"}),
		},
		Messages: []Message{
			{Role: "user", Message: "X"},
			{Role: "assistant", Message: "ans"},
		},
	})

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-new", output: "answer"},
		},
	}

	err := RunExplainWithRunner([]string{"A", "B"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 1 {
		t.Fatalf("expected 1 start call, got %d", runner.startCalled)
	}
	if runner.resumeCalled != 1 {
		t.Fatalf("expected 1 resume for follow-up (inner resume after new session), got %d", runner.resumeCalled)
	}
}

func TestRunExplain_ModelFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-model", output: "answer"},
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

func TestRunExplain_VerboseFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	runner := &mockRunner{
		startOutputs: []mockRunOutput{
			{sessionID: "sess-v", output: "verbose output"},
		},
	}

	err := RunExplainWithRunner([]string{"-v", "hello"}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.startCalled != 1 {
		t.Fatalf("expected 1 start call, got %d", runner.startCalled)
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
