## Expected

- One HTTP 200 response.
- `message.content` contains both `preset:think:think-message` and `preset:message:think-message` (think collapsed into reply).
- No `tool_calls`.
- `finish_reason` = `stop`.
- Agent-events: exactly 2 lines — `think` then `message` (both consumed on #1 serve).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	msg := choice["message"].(map[string]any)
	content, _ := msg["content"].(string)
	if content == "" {
		t.Fatalf("expected non-empty merged content, got %v", msg["content"])
	}
	assertContains(t, content, "preset:think:think-message")
	assertContains(t, content, "preset:message:think-message")
	if choice["finish_reason"] != "stop" {
		t.Fatalf("expected finish_reason=stop, got %q", choice["finish_reason"])
	}

	if len(resp.AgentEventsLines) != 2 {
		t.Fatalf("agent-events: want 2 lines (think+message on one serve), got %d\n%s",
			len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "think" || events[1]["type"] != "message" {
		t.Fatalf("want think then message, got %v then %v", events[0]["type"], events[1]["type"])
	}
}
```