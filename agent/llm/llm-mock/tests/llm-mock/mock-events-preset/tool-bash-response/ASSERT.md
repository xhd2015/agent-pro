---
label: e2e
---

## Expected

- HTTP 200.
- `message.content` is null.
- `message.tool_calls` has one entry with `function.name` = `bash`.
- `finish_reason` is `tool_calls`.
- Agent-events log has one `type` = `tool_call` line with `tool` = `bash`.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(resp.Responses))
	}
	r := resp.Responses[0]
	if r.StatusCode != 200 {
		t.Fatalf("expected 200, got %d\nbody: %s", r.StatusCode, r.Body)
	}

	obj := parseJSON(t, r.Body)
	choice := obj["choices"].([]any)[0].(map[string]any)
	message := choice["message"].(map[string]any)

	if message["content"] != nil {
		t.Fatalf("expected null content, got %v", message["content"])
	}

	toolCalls, ok := message["tool_calls"].([]any)
	if !ok || len(toolCalls) != 1 {
		t.Fatalf("expected 1 tool_call, got %v", message["tool_calls"])
	}
	tc := toolCalls[0].(map[string]any)
	fn := tc["function"].(map[string]any)
	if fn["name"] != "bash" {
		t.Fatalf("expected function.name=bash, got %q", fn["name"])
	}

	if choice["finish_reason"] != "tool_calls" {
		t.Fatalf("expected finish_reason=tool_calls, got %q", choice["finish_reason"])
	}

	if len(resp.AgentEventsLines) < 1 {
		t.Fatalf("agent-events: want tool_call line, got 0\ncontent:\n%s", resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "tool_call" {
		t.Fatalf("first agent-event want tool_call, got %v", events[0]["type"])
	}
	tool, _ := events[0]["tool"].(string)
	if tool != "bash" {
		t.Fatalf("agent-event tool want bash, got %q", tool)
	}
}
```