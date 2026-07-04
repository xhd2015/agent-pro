package mockgen

import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func TestAgentEventToExchangeResponseThink(t *testing.T) {
	resp := AgentEventToExchangeResponse(types.AgentEvent{
		Type: types.ActionThink,
		Text: "thinking about the task",
	})
	if resp.Content == nil || *resp.Content != "thinking about the task" {
		t.Fatalf("content = %v, want thinking about the task", resp.Content)
	}
	if resp.FinishReason != "stop" {
		t.Fatalf("finish_reason = %q, want stop", resp.FinishReason)
	}
	if resp.ToolCalls != nil {
		t.Fatalf("expected no tool_calls, got %v", resp.ToolCalls)
	}
}

func TestAgentEventToExchangeResponseMessage(t *testing.T) {
	resp := AgentEventToExchangeResponse(types.AgentEvent{
		Type: types.ActionMessage,
		Text: "done with the task",
	})
	if resp.Content == nil || *resp.Content != "done with the task" {
		t.Fatalf("content = %v, want done with the task", resp.Content)
	}
	if resp.FinishReason != "stop" {
		t.Fatalf("finish_reason = %q, want stop", resp.FinishReason)
	}
}

func TestAgentEventToExchangeResponseToolCall(t *testing.T) {
	resp := AgentEventToExchangeResponse(types.AgentEvent{
		ID:   "evt_42",
		Type: types.ActionToolCall,
		Tool: "bash",
		ToolInput: map[string]any{
			"command": "echo hello",
		},
	})
	if resp.Content != nil {
		t.Fatalf("content = %v, want nil", resp.Content)
	}
	if resp.FinishReason != "tool_calls" {
		t.Fatalf("finish_reason = %q, want tool_calls", resp.FinishReason)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("tool_calls = %v, want 1 entry", resp.ToolCalls)
	}
	tc := resp.ToolCalls[0]
	if tc.ID != "evt_42" || tc.Type != "function" || tc.Function.Name != "bash" {
		t.Fatalf("unexpected tool call: %+v", tc)
	}
	if tc.Function.Arguments != `{"command":"echo hello"}` {
		t.Fatalf("arguments = %q", tc.Function.Arguments)
	}
}

func TestSeedFromPromptDeterministic(t *testing.T) {
	t.Setenv("LLM_MOCK_RANDOM_SEED", "")
	a := SeedFromPrompt("same prompt")
	b := SeedFromPrompt("same prompt")
	if a != b {
		t.Fatalf("expected deterministic seed, got %d and %d", a, b)
	}
	if a == SeedFromPrompt("different prompt") {
		t.Fatal("expected different prompts to produce different seeds")
	}
}

func TestSeedFromPromptEnvOverride(t *testing.T) {
	t.Setenv("LLM_MOCK_RANDOM_SEED", "424242")
	if got := SeedFromPrompt("ignored"); got != 424242 {
		t.Fatalf("seed = %d, want 424242", got)
	}
}