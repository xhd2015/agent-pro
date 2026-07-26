---
label: e2e
---

## Expected

- Two HTTP 200 responses with non-empty `message.content`.
- Second response is generated fallback (not `no_match`, not preset replay).
- Agent-events: first line `type` = `message` (preset), second line `type` = `think` (genStream).

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

	var content1, content2 string
	for i, r := range resp.Responses {
		if r.StatusCode != 200 {
			t.Fatalf("response %d: expected 200, got %d\nbody: %s", i+1, r.StatusCode, r.Body)
		}
		obj := parseJSON(t, r.Body)
		if errObj, ok := obj["error"].(map[string]any); ok {
			if errObj["type"] == "no_match" {
				t.Fatalf("response %d: expected success, got no_match: %s", i+1, r.Body)
			}
		}
		msg := obj["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
		content, _ := msg["content"].(string)
		if content == "" {
			t.Fatalf("response %d: expected non-empty content, got %v", i+1, msg)
		}
		if i == 0 {
			content1 = content
		} else {
			content2 = content
		}
	}
	if content1 == content2 {
		t.Fatalf("expected preset message then genStream think (distinct content), got same %q", content1)
	}

	if len(resp.AgentEventsLines) < 2 {
		t.Fatalf("agent-events: want >=2 lines (message+think), got %d\ncontent:\n%s",
			len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "message" {
		t.Fatalf("first agent-event want message (preset), got %v\n%s", events[0]["type"], resp.AgentEventsContent)
	}
	thinkIdx := -1
	for i, ev := range events {
		if ev["type"] == "think" {
			thinkIdx = i
			break
		}
	}
	if thinkIdx < 0 {
		t.Fatalf("missing think AgentEvent from genStream fallback:\n%s", resp.AgentEventsContent)
	}
	if thinkIdx == 0 {
		t.Fatalf("genStream think must follow preset message dequeue; think at index 0\n%s",
			resp.AgentEventsContent)
	}
}
```