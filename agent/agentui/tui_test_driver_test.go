package agentui

import (
	"errors"
	"strings"
	"testing"
)

func TestTUITestDriverClarificationFlow(t *testing.T) {
	var calls []TUITestLLMCall
	driver := NewTUITestDriver(TUITestDriverOptions{
		Done:        true,
		Model:       "test-model",
		AgentRunner: "fake-codex",
		SessionID:   "session-1",
		SessionDir:  t.TempDir(),
		Questions: []TUITestQuestion{
			{ID: "q1", Question: "Which backend?", Options: []QuestionOption{{Label: "Codex"}, {Label: "OpenCode"}}},
		},
		OnLLMStart: func(call TUITestLLMCall) TUITestLLMResult {
			calls = append(calls, call)
			return TUITestLLMResult{Output: "finished"}
		},
	})

	driver.DeliverLLMDone("need clarification", nil)
	snap := driver.Snapshot()
	if !snap.ClarificationMode {
		t.Fatal("expected clarification mode")
	}
	if snap.SelectedQuestion != "q1" {
		t.Fatalf("SelectedQuestion = %q, want q1", snap.SelectedQuestion)
	}
	if !strings.Contains(snap.View, "Which backend?") {
		t.Fatalf("view missing question:\n%s", snap.View)
	}

	driver.Enter()
	if len(calls) != 1 {
		t.Fatalf("LLM calls = %d, want 1", len(calls))
	}
	if calls[0].AgentRunner != "fake-codex" {
		t.Fatalf("AgentRunner = %q, want fake-codex", calls[0].AgentRunner)
	}
	if !strings.Contains(calls[0].Prompt, "Codex") {
		t.Fatalf("resume prompt missing selected answer:\n%s", calls[0].Prompt)
	}
	if !driver.FlushLLM() {
		t.Fatal("expected queued LLM result")
	}
	snap = driver.Snapshot()
	if !snap.Done {
		t.Fatal("expected done after flushing LLM")
	}
	if snap.ClarificationMode {
		t.Fatal("expected clarification mode to end")
	}
	if snap.LLMOutput != "finished" {
		t.Fatalf("LLMOutput = %q, want finished", snap.LLMOutput)
	}
}

func TestTUITestDriverTypedAnswerOverridesOption(t *testing.T) {
	driver := NewTUITestDriver(TUITestDriverOptions{
		Done: true,
		Questions: []TUITestQuestion{
			{ID: "q1", Question: "Which backend?", Options: []QuestionOption{{Label: "Codex"}, {Label: "OpenCode"}}},
		},
		OnLLMStart: func(call TUITestLLMCall) TUITestLLMResult {
			return TUITestLLMResult{Output: "done"}
		},
	})
	driver.DeliverLLMDone("need clarification", nil)
	driver.TypeText("Custom backend")
	driver.Enter()

	snap := driver.Snapshot()
	if snap.Questions[0].Answer != "Custom backend" {
		t.Fatalf("answer = %q, want Custom backend", snap.Questions[0].Answer)
	}
}

func TestTUITestDriverCapturesLLMError(t *testing.T) {
	driver := NewTUITestDriver(TUITestDriverOptions{})
	driver.DeliverLLMDone("", errors.New("backend failed"))

	snap := driver.Snapshot()
	if !snap.Done {
		t.Fatal("expected done after error")
	}
	if snap.Error != "backend failed" {
		t.Fatalf("Error = %q, want backend failed", snap.Error)
	}
	if len(snap.Logs) == 0 || !strings.Contains(strings.Join(snap.Logs, "\n"), "backend failed") {
		t.Fatalf("logs missing backend error: %v", snap.Logs)
	}
}
