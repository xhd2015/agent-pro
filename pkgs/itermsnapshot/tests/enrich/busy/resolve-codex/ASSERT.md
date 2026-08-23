## Expected

- Capture succeeds.
- Exactly one Agent keyed by `sess-busy-codex`.
- Kind=`codex`, SessionID=`codex-sess-xyz`, Title empty.
- Tree non-empty with a codex role node.

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
		t.Fatalf("want 1 agent, got %d", agentCount(res.Agents))
	}
	ag := res.Agents["sess-busy-codex"]
	if ag == nil {
		t.Fatal("missing Agents[sess-busy-codex]")
	}
	if ag.Kind != "codex" {
		t.Fatalf("Kind=%q want codex", ag.Kind)
	}
	if ag.SessionID != "codex-sess-xyz" {
		t.Fatalf("SessionID=%q want codex-sess-xyz", ag.SessionID)
	}
	if ag.Title != "" {
		t.Fatalf("Title=%q want empty for codex", ag.Title)
	}
	if len(ag.Tree) < 1 {
		t.Fatal("expected non-empty Tree")
	}
	found := false
	for _, n := range ag.Tree {
		if n.Role == "codex" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Tree missing codex role: %+v", ag.Tree)
	}
}
```
