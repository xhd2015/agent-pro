## Expected

- Exit code 0.
- Log file at `LogEventsPath` exists with at least 2 JSONL lines.
- Each line is a standard `AgentEvent` with non-empty `type` (not RecordedRequest).
- At least 2 lines have `type` = `"message"`.
- Message `text` values collectively contain `from-config` and `from-events`.

## Side Effects

- No `--log-events` on subcommand argv; flag supplied only via `LLM_MOCK_RUN_FLAGS`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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