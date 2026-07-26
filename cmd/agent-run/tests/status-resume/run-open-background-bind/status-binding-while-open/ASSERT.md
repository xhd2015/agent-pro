---
label: e2e
---

## Expected

### Mid-open status probe (`status --json`)

- Probe exit 0.
- `runner.status` is `binding` (preferred while discover pending) or `bound`
  (if bind finished before the probe). Must **not** be only a silent `unbound`
  with no bind-in-progress signal once background bind has started for this open.

### After open completes

- Open exit 0.
- Stderr contains `grok-tty: grok session` with the delayed UUID.
- Meta for the fixed session has `runner_session_id` equal to the UUID.

## Exit Code

open: 0; status probe: 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("open/status failed: %v elapsed=%s\nstdout:\n%s\nstderr:\n%s\nprobe stdout:\n%s",
			err, resp.Elapsed, resp.Stdout, resp.Stderr, resp.StatusProbeStdout)
	}

	// --- mid-open status ---
	if resp.StatusProbeExit != 0 {
		t.Fatalf("status probe exit=%d want 0\nstdout:\n%s\nstderr:\n%s",
			resp.StatusProbeExit, resp.StatusProbeStdout, resp.StatusProbeStderr)
	}
	status := ""
	if resp.StatusProbeJSON != nil {
		if s, ok := jsonPathString(resp.StatusProbeJSON, "runner", "status"); ok {
			status = strings.ToLower(strings.TrimSpace(s))
		}
	}
	if status == "" {
		// Fall back to raw stdout: match runner.status without treating "unbound" as "bound".
		low := strings.ToLower(resp.StatusProbeStdout)
		switch {
		case strings.Contains(low, `"status": "binding"`) || strings.Contains(low, `"status":"binding"`):
			status = "binding"
		case strings.Contains(low, `"status": "unbound"`) || strings.Contains(low, `"status":"unbound"`):
			status = "unbound"
		case strings.Contains(low, `"status": "bound"`) || strings.Contains(low, `"status":"bound"`):
			status = "bound"
		}
	}
	// Contract: while open bind is in flight (or just finished), runner layer must
	// report binding or bound — plain unbound means no mid-open bind signal.
	if status != "binding" && status != "bound" {
		t.Fatalf("mid-open runner.status=%q want binding|bound (background bind in progress or done); probe stdout:\n%s\nopen stderr:\n%s",
			status, resp.StatusProbeStdout, resp.Stderr)
	}

	// --- open completion ---
	assertSuccess(t, resp)
	assertContains(t, resp.Stderr, "grok-tty: grok session "+bgBindStatusUUID)
	if resp.MetaAfter != nil {
		if id, _ := resp.MetaAfter["runner_session_id"].(string); strings.TrimSpace(id) != bgBindStatusUUID {
			t.Fatalf("meta.runner_session_id=%q want %q; meta=%v", id, bgBindStatusUUID, resp.MetaAfter)
		}
	} else if _, id, ok := findMetaRunnerSessionID(t, req.Home, "grok-tty", bgBindStatusUUID); !ok || id != bgBindStatusUUID {
		t.Fatalf("missing meta runner_session_id=%q after open; probe was %q", bgBindStatusUUID, status)
	}
}
```
