---
label: e2e
---

## Expected

- Exit code 0.
- Log file at `LogEventsPath` exists with at least 1 JSONL line.
- Each line is a standard `AgentEvent` with non-empty `type` (not RecordedRequest).
- Log contains at least one `message` AgentEvent.

## Side Effects

- `--log-events` emits `agent/event/types` `AgentEvent` JSONL when the mock **serves** a response.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.LogEventsLines) < 1 {
		t.Fatalf("log-events file: want >=1 JSONL line, got %d\ncontent:\n%s",
			len(resp.LogEventsLines), resp.LogEventsContent)
	}

	events, parseErr := parseAgentEventMaps(resp.LogEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	ok, missing := agentEventsHaveTypes(events, "message")
	if !ok {
		t.Fatalf("log-events missing AgentEvent types: %s\ncontent:\n%s", missing, resp.LogEventsContent)
	}
}
```