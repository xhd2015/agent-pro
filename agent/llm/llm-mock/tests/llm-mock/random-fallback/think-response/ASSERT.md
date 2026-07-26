---
label: e2e
---

## Expected

- HTTP 200.
- `choices[0].message.content` is a non-empty string (thought text from `ActionThink`).
- `choices[0].finish_reason` is `"stop"`.
- No `tool_calls` on the first generated event (think precedes tool/message events).

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
		t.Fatalf("expected status 200, got %d\nbody: %s", r.StatusCode, r.Body)
	}

	obj := parseJSON(t, r.Body)
	choice := obj["choices"].([]any)[0].(map[string]any)
	message := choice["message"].(map[string]any)

	content, ok := message["content"].(string)
	if !ok || content == "" {
		t.Fatalf("expected non-empty think content, got message=%v", message)
	}

	if choice["finish_reason"] != "stop" {
		t.Fatalf("expected finish_reason='stop' for think, got %q", choice["finish_reason"])
	}

	if message["tool_calls"] != nil {
		t.Fatalf("expected no tool_calls on first think event, got %v", message["tool_calls"])
	}
}
```