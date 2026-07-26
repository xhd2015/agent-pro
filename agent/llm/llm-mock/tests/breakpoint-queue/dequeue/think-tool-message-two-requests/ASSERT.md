---
label: e2e
---

## Expected

- Two HTTP 200 responses on `POST /v1/chat/completions`.
- Response #1: `message.content` is null; exactly one `tool_calls` entry with `function.name` = `bash`; `finish_reason` = `tool_calls`.
- Response #1 body must **not** contain preset message text `preset:message:think-tool-message`.
- Response #1 body must **not** contain preset think text `preset:think:think-tool-message` in `message.content` (think omitted from chat tool wire).
- Response #2: non-empty `message.content` containing `preset:message:think-tool-message`; no `tool_calls`.
- Agent-events: exactly 3 lines — `think`, `tool_call`, `message` in order.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(resp.Responses))
	}

	// Response #1: tool_call breakpoint
	r1 := resp.Responses[0]
	if r1.StatusCode != 200 {
		t.Fatalf("response 1: expected 200, got %d\nbody: %s", r1.StatusCode, r1.Body)
	}
	obj1 := parseJSON(t, r1.Body)
	choice1 := obj1["choices"].([]any)[0].(map[string]any)
	msg1 := choice1["message"].(map[string]any)
	if msg1["content"] != nil {
		t.Fatalf("response 1: expected null content for tool breakpoint, got %v", msg1["content"])
	}
	toolCalls, ok := msg1["tool_calls"].([]any)
	if !ok || len(toolCalls) != 1 {
		t.Fatalf("response 1: expected exactly 1 tool_call, got %v", msg1["tool_calls"])
	}
	tc := toolCalls[0].(map[string]any)
	fn := tc["function"].(map[string]any)
	if fn["name"] != "bash" {
		t.Fatalf("response 1: expected bash tool, got %q", fn["name"])
	}
	if choice1["finish_reason"] != "tool_calls" {
		t.Fatalf("response 1: expected finish_reason=tool_calls, got %q", choice1["finish_reason"])
	}
	assertNotContains(t, r1.Body, "preset:message:think-tool-message")
	assertNotContains(t, r1.Body, "preset:think:think-tool-message")

	// Response #2: message breakpoint
	r2 := resp.Responses[1]
	if r2.StatusCode != 200 {
		t.Fatalf("response 2: expected 200, got %d\nbody: %s", r2.StatusCode, r2.Body)
	}
	obj2 := parseJSON(t, r2.Body)
	msg2 := obj2["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
	content, _ := msg2["content"].(string)
	if content == "" {
		t.Fatalf("response 2: expected non-empty message content")
	}
	assertContains(t, content, "preset:message:think-tool-message")
	if _, hasTools := msg2["tool_calls"]; hasTools && msg2["tool_calls"] != nil {
		t.Fatalf("response 2: expected no tool_calls, got %v", msg2["tool_calls"])
	}

	// Agent-events consumption order
	if len(resp.AgentEventsLines) != 3 {
		t.Fatalf("agent-events: want 3 lines (think, tool_call, message), got %d\n%s",
			len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "think" || events[1]["type"] != "tool_call" || events[2]["type"] != "message" {
		t.Fatalf("want think, tool_call, message order, got %v, %v, %v",
			events[0]["type"], events[1]["type"], events[2]["type"])
	}
}
```