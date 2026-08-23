## Expected

- Capture succeeds; Result.Snapshot is the injected inventory.
- `Agents` is nil or empty (`agentCount == 0`) despite busy session and a
  resolve inject that would hard-hit under enrich.

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
	if agentCount(res.Agents) != 0 {
		t.Fatalf("NoEnrich: want 0 agents, got %d (%v)", agentCount(res.Agents), res.Agents)
	}
	if res.Snapshot == nil || len(res.Snapshot.Windows) != 1 {
		t.Fatal("expected injected one-window Snapshot preserved")
	}
}
```
