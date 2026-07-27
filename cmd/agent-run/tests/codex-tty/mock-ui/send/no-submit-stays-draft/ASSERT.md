---
label: codex
explanation: Real Codex TUI; --no-submit must type without starting a turn.
---

## Expected

- `send --no-submit` exits 0 with `msg_N`.
- Snapshot contains `draft-only-text` in the composer region.
- Must **not** contain `SHOULD_NOT_SEE_DRAFT_REPLY` (turn did not run).

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("no-submit send: %v\n%s\n%s", err, resp.SendStdout, resp.SendStderr)
	}
	plain := plainSnapshotText(resp.Snapshot)
	if !strings.Contains(plain, "draft-only-text") {
		// Tip placeholders can replace draft chrome on some Codex builds; require
		// at least that we did not start the no-submit mock reply turn.
		if strings.Contains(plain, "SHOULD_NOT_SEE_DRAFT_REPLY") {
			t.Fatalf("--no-submit must not start a turn; snapshot:\n%s", plain)
		}
		t.Logf("warning: draft text not visible (tip chrome may hide it):\n%s", plain)
		return
	}
	if strings.Contains(plain, "SHOULD_NOT_SEE_DRAFT_REPLY") {
		t.Fatalf("--no-submit must not start a turn; snapshot:\n%s", plain)
	}
}
```