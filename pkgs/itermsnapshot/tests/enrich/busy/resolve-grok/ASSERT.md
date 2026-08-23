## Expected

- Capture succeeds.
- Exactly one Agent keyed by `sess-busy-grok`.
- Kind=`grok`, SessionID=`abc-grok-session`, Title=`My Grok Title`.
- Tree has 2 nodes with Roles `input` then `grok`.

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
	if agentCount(res.Agents) != 1 {
		t.Fatalf("want 1 agent, got %d (%v)", agentCount(res.Agents), res.Agents)
	}
	ag := res.Agents["sess-busy-grok"]
	if ag == nil {
		t.Fatal("missing Agents[sess-busy-grok]")
	}
	if ag.Kind != "grok" {
		t.Fatalf("Kind=%q want grok", ag.Kind)
	}
	if ag.SessionID != "abc-grok-session" {
		t.Fatalf("SessionID=%q want abc-grok-session", ag.SessionID)
	}
	if ag.Title != "My Grok Title" {
		t.Fatalf("Title=%q want My Grok Title", ag.Title)
	}
	if len(ag.Tree) != 2 {
		t.Fatalf("Tree len=%d want 2", len(ag.Tree))
	}
	if ag.Tree[0].Role != "input" || ag.Tree[0].PID != 5100 {
		t.Fatalf("Tree[0]=%+v want input pid 5100", ag.Tree[0])
	}
	if ag.Tree[1].Role != "grok" || ag.Tree[1].Cmd != "grok" {
		t.Fatalf("Tree[1]=%+v want role=grok cmd=grok", ag.Tree[1])
	}
}
```
