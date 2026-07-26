---
label: e2e
---

## Expected

- Three HTTP 200 responses.
- Response #1: one `tool_calls` with `function.name` = `bash`; `finish_reason` = `tool_calls`; null content.
- Response #2: one `tool_calls` with `function.name` = `read`; `finish_reason` = `tool_calls`; null content.
- Response #3: non-empty `message.content` with `preset:message:two-tool-message`; no `tool_calls`.
- Agent-events: exactly 3 lines — two `tool_call` then one `message` (never two tools logged on one serve).

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(resp.Responses))
	}

	wantTools := []string{"bash", "read"}
	for i, wantTool := range wantTools {
		r := resp.Responses[i]
		if r.StatusCode != 200 {
			t.Fatalf("response %d: expected 200, got %d\nbody: %s", i+1, r.StatusCode, r.Body)
		}
		obj := parseJSON(t, r.Body)
		choice := obj["choices"].([]any)[0].(map[string]any)
		msg := choice["message"].(map[string]any)
		if msg["content"] != nil {
			t.Fatalf("response %d: expected null content, got %v", i+1, msg["content"])
		}
		toolCalls, ok := msg["tool_calls"].([]any)
		if !ok || len(toolCalls) != 1 {
			t.Fatalf("response %d: expected exactly 1 tool_call, got %v", i+1, msg["tool_calls"])
		}
		fn := toolCalls[0].(map[string]any)["function"].(map[string]any)
		if fn["name"] != wantTool {
			t.Fatalf("response %d: expected tool %q, got %q", i+1, wantTool, fn["name"])
		}
		if choice["finish_reason"] != "tool_calls" {
			t.Fatalf("response %d: expected finish_reason=tool_calls, got %q", i+1, choice["finish_reason"])
		}
	}

	r3 := resp.Responses[2]
	obj3 := parseJSON(t, r3.Body)
	msg3 := obj3["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
	content, _ := msg3["content"].(string)
	if content == "" {
		t.Fatalf("response 3: expected non-empty message content")
	}
	assertContains(t, content, "preset:message:two-tool-message")

	if len(resp.AgentEventsLines) != 3 {
		t.Fatalf("agent-events: want 3 lines, got %d\n%s", len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "tool_call" || events[1]["type"] != "tool_call" || events[2]["type"] != "message" {
		t.Fatalf("want tool_call, tool_call, message order, got %v, %v, %v",
			events[0]["type"], events[1]["type"], events[2]["type"])
	}
	tool0, _ := events[0]["tool"].(string)
	tool1, _ := events[1]["tool"].(string)
	if tool0 != "bash" || tool1 != "read" {
		t.Fatalf("want bash then read tools, got %q then %q", tool0, tool1)
	}
}
```