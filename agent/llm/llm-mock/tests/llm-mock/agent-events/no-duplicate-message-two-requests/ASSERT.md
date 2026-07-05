## Expected

- Both HTTP responses return 200.
- Agent-events log has **exactly 4** JSONL lines: two serves × (one `think`, one `message`) with breakpoint dequeue.
- Exactly **two** `message` events (one per serve, no peek-ahead duplicate within a serve).
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

	if len(resp.AgentEventsLines) != 4 {
		t.Fatalf("agent-events: want exactly 4 lines (2 serves × think+message), got %d\ncontent:\n%s",
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

	if thinkCount != 2 {
		t.Fatalf("want 2 think events (one per serve), got %d in:\n%s", thinkCount, resp.AgentEventsContent)
	}
	if messageCount != 2 {
		t.Fatalf("want 2 message events (one per serve, no duplicate), got %d in:\n%s", messageCount, resp.AgentEventsContent)
	}
	for i := 0; i < len(events); i += 2 {
		if events[i]["type"] != "think" || events[i+1]["type"] != "message" {
			t.Fatalf("want think then message per serve, got %v then %v at lines %d-%d\n%s",
				events[i]["type"], events[i+1]["type"], i+1, i+2, resp.AgentEventsContent)
		}
	}
}
```