---
label: e2e
---

## Expected

- Two HTTP 200 responses.
- First response `choices[0].message.content` is `from-prefix`.
- Second response is HTTP 200 with preset message content (not `from-prefix`, not `no_match`).
- Agent-events second line has `type` = `message` (preset dequeue).

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

	r1 := resp.Responses[0]
	if r1.StatusCode != 200 {
		t.Fatalf("response 1: expected 200, got %d\nbody: %s", r1.StatusCode, r1.Body)
	}
	obj1 := parseJSON(t, r1.Body)
	m1 := obj1["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
	if m1["content"] != "from-prefix" {
		t.Fatalf("response 1: expected from-prefix, got %q", m1["content"])
	}

	r2 := resp.Responses[1]
	if r2.StatusCode != 200 {
		t.Fatalf("response 2: expected 200 (preset simple), got %d\nbody: %s", r2.StatusCode, r2.Body)
	}
	obj2 := parseJSON(t, r2.Body)
	if errObj, ok := obj2["error"].(map[string]any); ok {
		if errObj["type"] == "no_match" {
			t.Fatalf("response 2: expected preset message, got no_match: %s", r2.Body)
		}
	}
	msg2 := obj2["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
	content2, _ := msg2["content"].(string)
	if content2 == "from-prefix" {
		t.Fatalf("response 2: expected preset message, got prefix replay %q", content2)
	}
	if content2 == "" {
		t.Fatalf("response 2: expected non-empty preset message content, got %v", msg2)
	}

	if len(resp.AgentEventsLines) < 1 {
		t.Fatalf("agent-events: want at least 1 preset message line, got 0\ncontent:\n%s",
			resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	foundMessage := false
	for _, ev := range events {
		if ev["type"] == "message" {
			foundMessage = true
			break
		}
	}
	if !foundMessage {
		t.Fatalf("agent-events: expected message AgentEvent from preset simple, got:\n%s",
			resp.AgentEventsContent)
	}
}
```