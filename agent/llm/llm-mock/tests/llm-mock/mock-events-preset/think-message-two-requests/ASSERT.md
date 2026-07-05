## Expected

- Two HTTP 200 responses with non-empty `message.content`.
- Response #1 merges preset think+message into one content (breakpoint dequeue).
- Response #2 serves genStream fallback (distinct non-empty content).
- Agent-events after both requests: at least 2 lines with `think` then `message` from preset serve on #1.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(resp.Responses))
	}

	var content1, content2 string
	for i, r := range resp.Responses {
		if r.StatusCode != 200 {
			t.Fatalf("response %d: expected 200, got %d\nbody: %s", i+1, r.StatusCode, r.Body)
		}
		obj := parseJSON(t, r.Body)
		msg := obj["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
		content, _ := msg["content"].(string)
		if content == "" {
			t.Fatalf("response %d: expected non-empty content, got %v", i+1, msg)
		}
		if i == 0 {
			content1 = content
		} else {
			content2 = content
		}
	}
	assertContains(t, content1, "preset:think:think-message")
	assertContains(t, content1, "preset:message:think-message")
	if content1 == content2 {
		t.Fatalf("expected distinct preset vs fallback content, got same %q", content1)
	}

	if len(resp.AgentEventsLines) < 2 {
		t.Fatalf("agent-events: want at least 2 lines from preset serve, got %d\ncontent:\n%s",
			len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "think" || events[1]["type"] != "message" {
		t.Fatalf("want think then message order on #1 serve, got %v then %v\n%s",
			events[0]["type"], events[1]["type"], resp.AgentEventsContent)
	}
	for i := 0; i < 2; i++ {
		text, _ := events[i]["text"].(string)
		if text == "" {
			t.Fatalf("line %d: AgentEvent missing text: %#v", i+1, events[i])
		}
	}
}
```