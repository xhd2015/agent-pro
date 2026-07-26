---
label: e2e
---

## Expected

- Exit code 0 within ExecTimeout (open + instant attach must not wait for hold sleep).
- Meta for `--session-id` has `runner=grok-tty` and `runner_session_id` = Grok UUID.
- Soft: stderr may include a post-attach `grok-tty: <id>` line (not required if
  product only prints after detach in some builds — meta bind is the hard check).

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed (expect --open + instant attach to finish before hold timeout): %v\nstdout:\n%s\nstderr:\n%s",
			err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	meta := readSessionMetaFile(t, req.Home, req.SessionID)
	gotRunner, _ := meta["runner"].(string)
	if gotRunner != "grok-tty" {
		t.Fatalf("meta.runner = %q, want grok-tty; meta=%v", gotRunner, meta)
	}
	gotRSID, _ := meta["runner_session_id"].(string)
	if gotRSID != req.GrokSessionID {
		t.Fatalf("meta.runner_session_id = %q, want %q; meta=%v", gotRSID, req.GrokSessionID, meta)
	}
}
```
