## Expected

- Fake codex exits 0 (curl succeeds).
- Combined output does not contain `no_match` error from mock.
- Log-events file has at least 3 AgentEvent lines with types `think`, `tool_call`, and `message`.
- Event order: `think` before `tool_call` before `message`.

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
	assertNotContains(t, combined, `"type":"no_match"`)
	assertNotContains(t, combined, `"type": "no_match"`)

	if len(resp.LogEventsLines) < 3 {
		t.Fatalf("log-events: want >=3 JSONL lines (think+tool_call+message), got %d\ncontent:\n%s",
			len(resp.LogEventsLines), resp.LogEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.LogEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	thinkIdx, toolIdx, messageIdx := -1, -1, -1
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
		case "tool_call":
			if toolIdx < 0 {
				toolIdx = i
			}
			tool, _ := ev["tool"].(string)
			if tool == "" {
				t.Fatalf("line %d: tool_call AgentEvent missing tool: %#v", i+1, ev)
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
	if toolIdx < 0 {
		t.Fatalf("missing tool_call AgentEvent in log:\n%s", resp.LogEventsContent)
	}
	if messageIdx < 0 {
		t.Fatalf("missing message AgentEvent in log:\n%s", resp.LogEventsContent)
	}
	if !(thinkIdx < toolIdx && toolIdx < messageIdx) {
		t.Fatalf("want think before tool_call before message; think@%d tool@%d message@%d\ncontent:\n%s",
			thinkIdx, toolIdx, messageIdx, resp.LogEventsContent)
	}
}
```