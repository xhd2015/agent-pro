## Expected

- Capture succeeds.
- Resolve was called with ShellPID `6100` (not zero / not skipped).
- Agent attached at `sess-shell-fallback` with Kind=grok.

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
	res := mustResult(t, resp, err)
	if req.ResolveCallPIDs == nil || len(*req.ResolveCallPIDs) != 1 {
		t.Fatalf("ResolveCallPIDs=%v want exactly [6100]", req.ResolveCallPIDs)
	}
	if (*req.ResolveCallPIDs)[0] != 6100 {
		t.Fatalf("resolve pid=%d want 6100 (ShellPID fallback)", (*req.ResolveCallPIDs)[0])
	}
	ag := res.Agents["sess-shell-fallback"]
	if ag == nil {
		t.Fatal("missing Agents[sess-shell-fallback]")
	}
	if ag.Kind != "grok" || ag.SessionID != "grok-via-shell" {
		t.Fatalf("agent=%+v want grok/grok-via-shell", ag)
	}
}
```
