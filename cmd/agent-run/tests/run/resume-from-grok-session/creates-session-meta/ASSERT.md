---
label: e2e
---

## Expected

- Exit code 0.
- `AGENT_RUN_HOME/sessions/<session-id>/meta.json` exists.
- `runner` is `grok-tty`.
- `runner_session_id` equals the imported Grok UUID.
- `session_id` matches the fixed `--session-id`.
- Prefer `workspace` equal to Grok `info.cwd` (absolute).

## Exit Code

0

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	meta := readSessionMetaFile(t, req.Home, req.SessionID)

	gotSID, _ := meta["session_id"].(string)
	if gotSID != req.SessionID {
		t.Fatalf("meta.session_id = %q, want %q; meta=%v", gotSID, req.SessionID, meta)
	}

	gotRunner, _ := meta["runner"].(string)
	if gotRunner != "grok-tty" {
		t.Fatalf("meta.runner = %q, want %q; meta=%v", gotRunner, "grok-tty", meta)
	}

	gotRSID, _ := meta["runner_session_id"].(string)
	if gotRSID != req.GrokSessionID {
		t.Fatalf("meta.runner_session_id = %q, want %q; meta=%v", gotRSID, req.GrokSessionID, meta)
	}

	// Workspace should match Grok cwd (import default; no --dir override).
	if ws, ok := meta["workspace"].(string); ok && strings.TrimSpace(ws) != "" {
		want := absPath(t, req.GrokCWD)
		got, err := filepath.Abs(ws)
		if err != nil {
			got = ws
		}
		if filepath.Clean(got) != filepath.Clean(want) {
			t.Fatalf("meta.workspace = %q, want Grok cwd %q", got, want)
		}
	} else {
		t.Fatalf("meta.workspace missing or empty; meta=%v", meta)
	}
}
```
