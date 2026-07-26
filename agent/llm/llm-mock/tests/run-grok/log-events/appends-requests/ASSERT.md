---
label: e2e
---

## Expected

- Exit code 0.
- Log file at `LogEventsPath` exists with at least 2 JSONL lines.
- Each line is a standard `AgentEvent` with non-empty `type` (not RecordedRequest).
- Lines must **not** have top-level `method` or `path` keys.
- At least 2 lines have `type` = `"message"`.
- Message `text` values collectively contain `from-config` and `from-events`.

## Side Effects

- Spec correction: prior test expected HTTP `RecordedRequest` bodies; `--log-events` must emit `agent/event/types` `AgentEvent` JSONL when the mock **serves** a response.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.LogEventsLines) < 2 {
		t.Fatalf("log-events file: want >=2 JSONL lines, got %d\ncontent:\n%s",
			len(resp.LogEventsLines), resp.LogEventsContent)
	}

	events, parseErr := parseAgentEventMaps(resp.LogEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	var messageTexts []string
	for i, ev := range events {
		typ, _ := ev["type"].(string)
		if typ == "" {
			t.Fatalf("line %d: missing type in %#v", i+1, ev)
		}
		if _, hasMethod := ev["method"]; hasMethod {
			t.Fatalf("line %d: RecordedRequest method key in AgentEvent log: %#v", i+1, ev)
		}
		if _, hasPath := ev["path"]; hasPath {
			t.Fatalf("line %d: RecordedRequest path key in AgentEvent log: %#v", i+1, ev)
		}
		if typ == "message" {
			text, _ := ev["text"].(string)
			messageTexts = append(messageTexts, text)
		}
	}

	if len(messageTexts) < 2 {
		t.Fatalf("want >=2 message AgentEvents, got %d in %#v", len(messageTexts), events)
	}

	combined := strings.Join(messageTexts, "\n")
	assertContains(t, combined, "from-config")
	assertContains(t, combined, "from-events")
}
```