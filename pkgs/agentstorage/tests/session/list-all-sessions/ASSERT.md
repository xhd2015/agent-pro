## Expected

- `ListSessions` returns exactly two sessions.
- Both `sess_opencode` (runner fake-opencode) and `sess_codex` (runner fake-codex) appear.
- Each item has matching SessionID and Runner fields.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if len(resp.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d: %+v", len(resp.Sessions), resp.Sessions)
	}
	byID := map[string]agentstorage.SessionMeta{}
	for _, s := range resp.Sessions {
		byID[s.SessionID] = s
	}
	s1, ok1 := byID[req.SessionID]
	s2, ok2 := byID[req.OtherSessID]
	if !ok1 || !ok2 {
		t.Fatalf("expected ids %q and %q in list, got %+v", req.SessionID, req.OtherSessID, resp.Sessions)
	}
	assertEqual(t, "s1.Runner", s1.Runner, req.Runner)
	assertEqual(t, "s2.Runner", s2.Runner, req.OtherRunner)
}
```
