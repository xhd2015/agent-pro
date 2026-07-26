---
label: e2e
---

## Expected

- Fake opencode exits 0 (curl succeeds).
- Combined output does not contain `no_match` error from mock.
- Log-events file has at least 1 AgentEvent line with type `message`.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	combined := resp.Stdout + resp.Stderr
	assertNotContains(t, combined, `"type":"no_match"`)
	assertNotContains(t, combined, `"type": "no_match"`)

	if len(resp.LogEventsLines) < 1 {
		t.Fatalf("log-events: want >=1 JSONL line (message), got %d\ncontent:\n%s",
			len(resp.LogEventsLines), resp.LogEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.LogEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	messageIdx := -1
	for i, ev := range events {
		typ, _ := ev["type"].(string)
		if typ == "message" {
			if messageIdx < 0 {
				messageIdx = i
			}
			text, _ := ev["text"].(string)
			if text == "" {
				t.Fatalf("line %d: message AgentEvent missing text: %#v", i+1, ev)
			}
		}
	}
	if messageIdx < 0 {
		t.Fatalf("missing message AgentEvent in log:\n%s", resp.LogEventsContent)
	}
}
```