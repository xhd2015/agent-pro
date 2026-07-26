---
label: e2e
---

## Expected

- Exit code 0.
- Log file at `LogEventsPath` exists with at least 2 JSONL lines.
- Each line is a standard `AgentEvent` (non-empty `type`; no top-level `method`/`path`).
- Log contains a `type` = `"think"` line before a `type` = `"message"` line.
- Think and message events have non-empty `text`.

## Side Effects

- Random fallback `findGeneratedMatch` must log raw `types.AgentEvent` objects before OpenAI conversion, not HTTP RecordedRequest bodies.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.LogEventsLines) < 2 {
		t.Fatalf("log-events file: want >=2 JSONL lines (think+message), got %d\ncontent:\n%s",
			len(resp.LogEventsLines), resp.LogEventsContent)
	}

	events, parseErr := parseAgentEventMaps(resp.LogEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	thinkIdx, messageIdx := -1, -1
	for i, ev := range events {
		typ, _ := ev["type"].(string)
		switch typ {
		case "think":
			if thinkIdx < 0 {
				thinkIdx = i
			}
			text, _ := ev["text"].(string)
			if text == "" {
				t.Fatalf("line %d: think AgentEvent missing text: %#v", i+1, ev)
			}
		case "message":
			if messageIdx < 0 {
				messageIdx = i
			}
			text, _ := ev["text"].(string)
			if text == "" {
				t.Fatalf("line %d: message AgentEvent missing text: %#v", i+1, ev)
			}
		}
	}

	if thinkIdx < 0 {
		t.Fatalf("missing think AgentEvent in log:\n%s", resp.LogEventsContent)
	}
	if messageIdx < 0 {
		t.Fatalf("missing message AgentEvent in log:\n%s", resp.LogEventsContent)
	}
	if thinkIdx >= messageIdx {
		t.Fatalf("want think before message; think@%d message@%d\ncontent:\n%s",
			thinkIdx, messageIdx, resp.LogEventsContent)
	}
}
```