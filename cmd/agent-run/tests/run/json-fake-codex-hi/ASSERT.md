---
label: e2e
---

## Expected

- Exit code 0.
- Stdout is NDJSON: every non-empty line is valid JSON `types.AgentEvent`.
- At least one event; the last event has `"type":"done"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	lines := parseJSONLines(t, resp.Stdout)
	if len(lines) == 0 {
		t.Fatal("expected at least one NDJSON event line on stdout")
	}
	last := lines[len(lines)-1]
	if last["type"] != "done" {
		t.Fatalf("expected last event type done, got %v", last["type"])
	}
}
```