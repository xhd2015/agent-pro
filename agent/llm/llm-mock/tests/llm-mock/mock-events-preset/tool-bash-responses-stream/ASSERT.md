## Expected

- HTTP 200 on `POST /v1/responses`.
- SSE body contains `function_call` output item (Responses API tool call), not only empty `output_text`.
- SSE body contains tool name `bash` and arguments referencing `preset-bash`.
- Agent-events log has one `type` = `tool_call` with `tool` = `bash`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
		t.Fatalf("expected function_call in Responses SSE, got empty-message stream only:\n%s", body)
	}
	if !strings.Contains(body, "bash") {
		t.Fatalf("expected bash tool name in Responses SSE:\n%s", body)
	}
	if !strings.Contains(body, "preset-bash") {
		t.Fatalf("expected preset-bash in function_call arguments:\n%s", body)
	}

	if len(resp.AgentEventsLines) < 1 {
		t.Fatalf("agent-events: want tool_call line, got 0\ncontent:\n%s", resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "tool_call" {
		t.Fatalf("first agent-event want tool_call, got %v", events[0]["type"])
	}
}
```