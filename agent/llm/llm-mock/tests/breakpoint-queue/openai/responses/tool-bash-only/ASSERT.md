---
label: e2e
---

## Expected

- HTTP 200 on `POST /v1/responses` with `stream=true`.
- SSE body contains `function_call` with tool name `bash` (or remapped `exec_command`).
- SSE body contains arguments referencing `preset-bash`.
- SSE body must **not** contain a reasoning item (no prefix think in queue).
- Agent-events: one `tool_call` line.

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

	if len(resp.Responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(resp.Responses))
	}
	r := resp.Responses[0]
	if r.StatusCode != 200 {
		t.Fatalf("expected 200, got %d\nbody: %s", r.StatusCode, r.Body)
	}

	body := r.Body
	if !strings.Contains(body, "function_call") {
		t.Fatalf("expected function_call in Responses SSE:\n%s", body)
	}
	if !strings.Contains(body, "bash") && !strings.Contains(body, "exec_command") {
		t.Fatalf("expected bash or exec_command in SSE:\n%s", body)
	}
	assertContains(t, body, "preset-bash")

	// No prefix think → no reasoning output item (ignore usage.reasoning_tokens metadata)
	if strings.Contains(body, `"type":"reasoning"`) || strings.Contains(body, "preset:think:") {
		t.Fatalf("tool-only breakpoint should not emit reasoning output item:\n%s", body)
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
}
```