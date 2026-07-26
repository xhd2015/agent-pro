---
label: e2e
---

## Expected

- Exit code 0.
- `AGENT_RUN_HOME/sessions/grok-tty/<id>/events.jsonl` exists.
- Persisted events include **user** `message`, **tool_call**, **assistant** `message`, and
  **think** (full-fidelity v1 mapping).
- Events are not a single end-of-run assistant blob only.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	path, lines := findGrokTTYEventsJSONL(t, req.Home)
	if path == "" {
		t.Fatal("events.jsonl not found under sessions/grok-tty/")
	}
	types := eventsCollectTypes(lines)
	if !types["message:user"] {
		t.Fatalf("events.jsonl missing user message event:\n%s", strings.Join(lines, "\n"))
	}
	if !types["tool_call"] {
		t.Fatalf("events.jsonl missing tool_call event:\n%s", strings.Join(lines, "\n"))
	}
	if !types["message:assistant"] {
		t.Fatalf("events.jsonl missing assistant message event:\n%s", strings.Join(lines, "\n"))
	}
	if !types["think"] {
		t.Fatalf("events.jsonl missing think event (full fidelity):\n%s", strings.Join(lines, "\n"))
	}
	if len(lines) < 3 {
		t.Fatalf("expected multiple streamed events, got %d lines:\n%s", len(lines), strings.Join(lines, "\n"))
	}
}
```