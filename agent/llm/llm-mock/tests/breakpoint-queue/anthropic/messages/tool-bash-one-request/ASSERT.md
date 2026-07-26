---
label: e2e
---

## Expected

- HTTP 200 on `POST /v1/messages`.
- `content[]` has exactly one `{type:"tool_use"}` with `name` = `bash`.
- No `{type:"thinking"}` blocks (no prefix think).
- Agent-events: one `tool_call` line with `tool` = `bash`.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(resp.Responses))
	}
	r := resp.Responses[0]
	if r.StatusCode != 200 {
		t.Fatalf("expected 200, got %d\nbody: %s", r.StatusCode, r.Body)
	}

	blocks := anthropicContentBlocks(t, r.Body)
	var toolUseCount, thinkingCount int
	for _, b := range blocks {
		typ, _ := b["type"].(string)
		switch typ {
		case "tool_use":
			toolUseCount++
			name, _ := b["name"].(string)
			if name != "bash" {
				t.Fatalf("expected tool_use name=bash, got %q", name)
			}
		case "thinking":
			thinkingCount++
		}
	}
	if toolUseCount != 1 {
		t.Fatalf("expected exactly 1 tool_use block, got %d in %v", toolUseCount, blocks)
	}
	if thinkingCount > 0 {
		t.Fatalf("tool-only breakpoint should not emit thinking blocks, got %d", thinkingCount)
	}

	if len(resp.AgentEventsLines) != 1 {
		t.Fatalf("agent-events: want 1 tool_call, got %d\n%s", len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "tool_call" {
		t.Fatalf("want tool_call, got %v", events[0]["type"])
	}
	tool, _ := events[0]["tool"].(string)
	if tool != "bash" {
		t.Fatalf("want bash tool, got %q", tool)
	}
}
```