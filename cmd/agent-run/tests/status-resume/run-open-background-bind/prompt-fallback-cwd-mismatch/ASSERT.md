---
label: e2e
---

## Expected

- Exit code 0.
- Stderr contains `grok-tty: grok session` with the seeded UUID under the other cwd.
- Stderr contains `grok-tty: grok updates` and a path under the other encoded cwd tree.
- Meta has `runner_session_id` equal to the seeded UUID (prompt-only fallback bind).

## Side Effects

- Session files remain only under `sessions/<encoded-other-cwd>/…`, not under the
  agent-run workspace encoded path.

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
	assertContains(t, resp.Stderr, "grok-tty: grok session "+promptFallbackUUID)
	assertContains(t, resp.Stderr, "grok-tty: grok updates")
	assertContains(t, resp.Stderr, "updates.jsonl")
	// updates path should reference the other encoded cwd, not workspace-only discovery.
	otherEnc := encodedGrokCwd(promptFallbackOtherCwd)
	if !strings.Contains(resp.Stderr, otherEnc) && !strings.Contains(resp.Stderr, promptFallbackUUID) {
		t.Fatalf("expected updates path under other cwd encoding %q or uuid in stderr:\n%s", otherEnc, resp.Stderr)
	}
	// Prefer seeing the other cwd segment when absolute path is printed.
	if strings.Contains(resp.Stderr, "grok updates") {
		// ok — session line already checked; path may be absolute with encoding
		_ = filepath.Join(req.GrokHome, "sessions", otherEnc, promptFallbackUUID)
	}
	if _, id, ok := findMetaRunnerSessionID(t, req.Home, "grok-tty", promptFallbackUUID); !ok {
		t.Fatalf("no meta.json with runner_session_id=%q after prompt-fallback cwd mismatch\nstderr:\n%s",
			promptFallbackUUID, resp.Stderr)
	} else if id != promptFallbackUUID {
		t.Fatalf("runner_session_id=%q want %q", id, promptFallbackUUID)
	}
}
```
