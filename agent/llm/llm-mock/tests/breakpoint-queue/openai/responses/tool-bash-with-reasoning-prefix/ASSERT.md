---
label: e2e
---

## Expected

- HTTP 200 on streaming `/v1/responses`.
- SSE contains **reasoning** item with `preset:think:think-tool-message`.
- SSE contains `function_call` before any message `output_text` for the unconsumed message breakpoint.
- Tool name on wire is `bash` or codex-remapped `exec_command`; arguments reference `preset-inline-bash` or `preset-bash`.
- Agent-events: `think` and `tool_call` on #1 serve.

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
	if !strings.Contains(body, "reasoning") {
		t.Fatalf("expected reasoning output item in SSE:\n%s", body)
	}
	assertContains(t, body, "preset:think:think-tool-message")
	assertContains(t, body, "function_call")
	if !strings.Contains(body, "bash") && !strings.Contains(body, "exec_command") {
		t.Fatalf("expected bash or exec_command tool name in SSE:\n%s", body)
	}
	if !strings.Contains(body, "preset-inline-bash") && !strings.Contains(body, "preset-bash") {
		t.Fatalf("expected preset bash args in SSE:\n%s", body)
	}
	assertNotContains(t, body, "preset:message:think-tool-message")

	if len(resp.AgentEventsLines) != 2 {
		t.Fatalf("agent-events: want think+tool_call on #1, got %d\n%s",
			len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "think" || events[1]["type"] != "tool_call" {
		t.Fatalf("want think then tool_call, got %v then %v", events[0]["type"], events[1]["type"])
	}
}
```