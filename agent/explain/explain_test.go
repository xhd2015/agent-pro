package explain

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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

func TestExplainHelpListsSupportedRunners(t *testing.T) {
	t.Parallel()
	for _, runner := range []string{"opencode", "codex", "grok", "commandcode"} {
		if !strings.Contains(explainHelp, runner) {
			t.Fatalf("help missing runner %q:\n%s", runner, explainHelp)
		}
	}
	if !strings.Contains(explainHelp, "--agent-runner") {
		t.Fatalf("help missing --agent-runner:\n%s", explainHelp)
	}
}

func TestEncodeDecodeRunnerMeta(t *testing.T) {
	t.Parallel()

	opencodeMeta := encodeRunnerMeta("opencode", "oc-1")
	id, err := decodeRunnerSessionID("opencode", opencodeMeta)
	if err != nil {
		t.Fatalf("decode opencode: %v", err)
	}
	if id != "oc-1" {
		t.Fatalf("opencode session = %q, want oc-1", id)
	}

	codexMeta := encodeRunnerMeta("codex", "thr-1")
	id, err = decodeRunnerSessionID("codex", codexMeta)
	if err != nil {
		t.Fatalf("decode codex: %v", err)
	}
	if id != "thr-1" {
		t.Fatalf("codex session = %q, want thr-1", id)
	}

	// Backward-compat: accept session_id for codex too.
	id, err = decodeRunnerSessionID("codex", mustMarshalJSON(map[string]string{"session_id": "thr-alt"}))
	if err != nil {
		t.Fatalf("decode codex session_id: %v", err)
	}
	if id != "thr-alt" {
		t.Fatalf("codex alt session = %q, want thr-alt", id)
	}
}

func TestRunExplain_SupportedRunnersStoreAgentRunner(t *testing.T) {
	runners := []string{"opencode", "codex", "grok", "commandcode"}
	for _, wantRunner := range runners {
		t.Run(wantRunner, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv(debugConfigHomeEnv, "") // ensure default HOME-based layout

			mock := &mockRunner{
				startOutputs: []mockRunOutput{{sessionID: "sess-" + wantRunner, output: "ok"}},
			}
			args := []string{"hello-" + wantRunner}
			if wantRunner != "opencode" {
				args = []string{"--agent-runner", wantRunner, "hello-" + wantRunner}
			}
			if err := RunExplainWithRunner(args, mock); err != nil {
				t.Fatalf("RunExplainWithRunner: %v", err)
			}

			base := filepath.Join(home, defaultSessionsBaseDir, "sessions")
			entries, err := os.ReadDir(base)
			if err != nil {
				t.Fatalf("readdir sessions: %v", err)
			}
			if len(entries) != 1 {
				t.Fatalf("session dirs = %d, want 1", len(entries))
			}
			data, err := readSession(filepath.Join(base, entries[0].Name()))
			if err != nil {
				t.Fatalf("readSession: %v", err)
			}
			if data.AgentRunner != wantRunner {
				t.Fatalf("AgentRunner = %q, want %q", data.AgentRunner, wantRunner)
			}
			meta, ok := data.AgentRunnersMeta[wantRunner]
			if !ok || len(meta) == 0 {
				t.Fatalf("missing meta for %s", wantRunner)
			}
			gotID, err := decodeRunnerSessionID(wantRunner, meta)
			if err != nil {
				t.Fatalf("decode meta: %v", err)
			}
			if gotID != "sess-"+wantRunner {
				t.Fatalf("session id = %q, want sess-%s", gotID, wantRunner)
			}
		})
	}
}
