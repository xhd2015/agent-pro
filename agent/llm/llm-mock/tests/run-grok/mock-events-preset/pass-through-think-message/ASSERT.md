## Expected

- Fake grok exits 0 (both curls succeed).
- Combined output contains `R1=` and `R2=` with JSON bodies (no `no_match`).
- Log-events file has at least 2 AgentEvent lines: `think` before `message`.
- Both think and message events have non-empty `text`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	combined := resp.Stdout + resp.Stderr
	for _, label := range []string{"R1=", "R2="} {
		if !strings.Contains(combined, label) {
			t.Fatalf("missing %s in output:\n%s", label, combined)
		}
	}
	assertNotContains(t, combined, `"type":"no_match"`)
	assertNotContains(t, combined, `"type": "no_match"`)

	if len(resp.LogEventsLines) < 2 {
		t.Fatalf("log-events: want >=2 JSONL lines (think+message), got %d\ncontent:\n%s",
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