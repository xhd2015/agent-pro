## Expected

- Capture succeeds.
- Exactly two Agents keyed by `sess-multi-a` and `sess-multi-b`.
- A: Kind=grok, SessionID=grok-a, Title=Agent A.
- B: Kind=codex, SessionID=codex-b.
- No cross-wiring of session IDs.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	res := mustResult(t, resp, err)
	if agentCount(res.Agents) != 2 {
		t.Fatalf("want 2 agents, got %d (%v)", agentCount(res.Agents), res.Agents)
	}
	a := res.Agents["sess-multi-a"]
	b := res.Agents["sess-multi-b"]
	if a == nil || b == nil {
		t.Fatalf("missing keys: a=%v b=%v map=%v", a, b, res.Agents)
	}
	if a.Kind != "grok" || a.SessionID != "grok-a" || a.Title != "Agent A" {
		t.Fatalf("agent A=%+v want grok/grok-a/Agent A", a)
	}
	if b.Kind != "codex" || b.SessionID != "codex-b" {
		t.Fatalf("agent B=%+v want codex/codex-b", b)
	}
	// Independence: no shared pointer / swapped IDs
	if a.SessionID == b.SessionID {
		t.Fatal("session IDs must differ across independent agents")
	}
}
```
