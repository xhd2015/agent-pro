---
label: e2e
---

## Expected

- HTTP 200 with generated think content.
- Agent-events think line text contains `Hello` (the user's actual message).
- Agent-events think line text must **not** contain `<user_query>` as the topic placeholder.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 1 || resp.Responses[0].StatusCode != 200 {
		t.Fatalf("want HTTP 200, got %#v", resp.Responses)
	}

	if len(resp.AgentEventsLines) < 1 {
		t.Fatalf("want >=1 agent-events line, got 0\n%s", resp.AgentEventsContent)
	}

	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	var thinkText string
	for _, ev := range events {
		if ev["type"] == "think" {
			thinkText, _ = ev["text"].(string)
			break
		}
	}
	if thinkText == "" {
		t.Fatalf("missing think AgentEvent in:\n%s", resp.AgentEventsContent)
	}
	assertContains(t, thinkText, "Hello")
	assertNotContains(t, thinkText, "<user_query>")
}
```