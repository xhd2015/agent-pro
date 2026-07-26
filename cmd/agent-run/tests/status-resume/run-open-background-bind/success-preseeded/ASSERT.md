---
label: e2e
---

## Expected Output

After open completes, stderr includes:

```text
grok-tty: <terminal-or-session-id>
grok-tty: grok session 550e8400-e29b-41d4-a716-446655440801
grok-tty: grok updates <path>/updates.jsonl
```

## Expected

- Exit code 0.
- Stderr contains `grok-tty: grok session` with the preseeded UUID.
- Stderr contains `grok-tty: grok updates` and `updates.jsonl`.
- Some `sessions/grok-tty/*/meta.json` has `runner_session_id` equal to the UUID.

## Exit Code

0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)
	assertContains(t, resp.Stderr, "grok-tty: grok session "+bgBindPreseedUUID)
	assertContains(t, resp.Stderr, "grok-tty: grok updates")
	assertContains(t, resp.Stderr, "updates.jsonl")
	if _, id, ok := findMetaRunnerSessionID(t, req.Home, "grok-tty", bgBindPreseedUUID); !ok {
		t.Fatalf("no meta.json with runner_session_id=%q under home\nstderr:\n%s", bgBindPreseedUUID, resp.Stderr)
	} else if id != bgBindPreseedUUID {
		t.Fatalf("runner_session_id=%q want %q", id, bgBindPreseedUUID)
	}
}
```
