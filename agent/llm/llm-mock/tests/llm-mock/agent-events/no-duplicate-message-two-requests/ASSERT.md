## Expected

- Both HTTP responses return 200.
- Agent-events log has **exactly 2** JSONL lines for this turn: one `think`, one `message`.
- Exactly **one** `message` event (no duplicate message from peek-ahead logging).
- Every `id` field matches `evt_<digits>` (valid JSON, no embedded control characters).

## Exit Code

0

```go
import (
	"regexp"
	"testing"
)

var evtIDPattern = regexp.MustCompile(`^evt_\d+$`)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 2 {
		t.Fatalf("expected 2 HTTP responses, got %d", len(resp.Responses))
	}
	for i, r := range resp.Responses {
		if r.StatusCode != 200 {
			t.Fatalf("response %d: want 200, got %d body=%s", i+1, r.StatusCode, r.Body)
		}
	}

	if len(resp.AgentEventsLines) != 2 {
		t.Fatalf("agent-events: want exactly 2 lines (think+message), got %d\ncontent:\n%s",
			len(resp.AgentEventsLines), resp.AgentEventsContent)
	}

	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	thinkCount, messageCount := 0, 0
	for i, ev := range events {
		id, _ := ev["id"].(string)
		if !evtIDPattern.MatchString(id) {
			t.Fatalf("line %d: invalid AgentEvent id %q (want evt_<digits>)\n%#v", i+1, id, ev)
		}
		switch ev["type"] {
		case "think":
			thinkCount++
		case "message":
			messageCount++
		}
	}

	if thinkCount != 1 {
		t.Fatalf("want 1 think event, got %d in:\n%s", thinkCount, resp.AgentEventsContent)
	}
	if messageCount != 1 {
		t.Fatalf("want 1 message event (no duplicate), got %d in:\n%s", messageCount, resp.AgentEventsContent)
	}
	if events[0]["type"] != "think" || events[1]["type"] != "message" {
		t.Fatalf("want think then message order, got types %v %v\n%s",
			events[0]["type"], events[1]["type"], resp.AgentEventsContent)
	}
}
```