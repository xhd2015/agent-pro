---
label: e2e
---

## Expected

- HTTP 200 on `POST /v1/responses` with `stream=true`.
- SSE body contains a **reasoning** output item with think text `preset:think:think-tool-message`.
- SSE body contains `function_call` with tool name `bash`.
- SSE body must **not** contain message text `preset:message:think-tool-message` (message not consumed on #1).
- Agent-events on #1 serve: `think` and `tool_call` only (2 lines so far).

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
	// Reasoning item (option B): collapsed think text on Responses wire
	if !strings.Contains(body, "reasoning") {
		t.Fatalf("expected reasoning item in Responses SSE, got:\n%s", body)
	}
	assertContains(t, body, "preset:think:think-tool-message")
	assertContains(t, body, "function_call")
	assertContains(t, body, "bash")
	assertNotContains(t, body, "preset:message:think-tool-message")

	if len(resp.AgentEventsLines) != 2 {
		t.Fatalf("agent-events after #1: want 2 lines (think+tool_call), got %d\n%s",
			len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "think" || events[1]["type"] != "tool_call" {
		t.Fatalf("want think then tool_call on #1 serve, got %v then %v",
			events[0]["type"], events[1]["type"])
	}
}
```