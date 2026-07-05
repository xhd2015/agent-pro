## Expected

- Two HTTP 200 responses on `POST /v1/responses`.
- Response #1: contains `function_call` / `bash` (tool breakpoint); no message text `preset:message:think-tool-message`.
- Response #2: contains message text `preset:message:think-tool-message`; no `function_call`.
- Agent-events: 3 lines total — `think`, `tool_call`, `message` in order.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(resp.Responses))
	}

	r1 := resp.Responses[0]
	if r1.StatusCode != 200 {
		t.Fatalf("response 1: expected 200, got %d\nbody: %s", r1.StatusCode, r1.Body)
	}
	assertContains(t, r1.Body, "function_call")
	assertContains(t, r1.Body, "bash")
	assertNotContains(t, r1.Body, "preset:message:think-tool-message")

	r2 := resp.Responses[1]
	if r2.StatusCode != 200 {
		t.Fatalf("response 2: expected 200, got %d\nbody: %s", r2.StatusCode, r2.Body)
	}
	assertContains(t, r2.Body, "preset:message:think-tool-message")
	if strings.Contains(r2.Body, "function_call") {
		t.Fatalf("response 2: expected message only, found function_call:\n%s", r2.Body)
	}

	if len(resp.AgentEventsLines) != 3 {
		t.Fatalf("agent-events: want 3 lines, got %d\n%s", len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "think" || events[1]["type"] != "tool_call" || events[2]["type"] != "message" {
		t.Fatalf("want think, tool_call, message order, got %v, %v, %v",
			events[0]["type"], events[1]["type"], events[2]["type"])
	}
}
```