---
label: e2e
---

## Expected

- Process wall time ≥ materialization delay (proves hard wait, not soft 750ms exit).
- Exit code 0.
- Stderr contains `grok-tty: grok session` with the delayed UUID.
- Stderr contains `grok-tty: grok updates` / `updates.jsonl`.
- Meta under home has `runner_session_id` equal to the delayed UUID.
- Child process must not have relied on `GROK_HOME` env (leaf sets `NoGrokHomeEnv`).

## Side Effects

- Delayed session is written only under isolated `$HOME/.grok` after the delay.

## Exit Code

0

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v elapsed=%s\nstdout:\n%s\nstderr:\n%s", err, resp.Elapsed, resp.Stdout, resp.Stderr)
	}
	// Critical: must wait past soft 750ms and past materialization delay.
	minWait := hardRequireNoGrokHomeDelay - 200*time.Millisecond
	if minWait < time.Second {
		minWait = time.Second
	}
	if resp.Elapsed < minWait {
		t.Fatalf("open exited too quickly (elapsed=%s < minWait=%s); expected hard wait without GROK_HOME env\nstderr:\n%s",
			resp.Elapsed, minWait, resp.Stderr)
	}
	// Soft path would fail-fast well under 1.5s when materialization is delayed 2s.
	if resp.Elapsed < 1500*time.Millisecond {
		t.Fatalf("elapsed=%s suggests soft 750ms path; want hard require via non-empty prompt\nstderr:\n%s",
			resp.Elapsed, resp.Stderr)
	}
	assertSuccess(t, resp)
	assertContains(t, resp.Stderr, "grok-tty: grok session "+hardRequireNoGrokHomeUUID)
	assertContains(t, resp.Stderr, "grok-tty: grok updates")
	assertContains(t, resp.Stderr, "updates.jsonl")
	if _, id, ok := findMetaRunnerSessionID(t, req.Home, "grok-tty", hardRequireNoGrokHomeUUID); !ok {
		t.Fatalf("no meta.json with runner_session_id=%q after hard wait without GROK_HOME (elapsed=%s)\nstderr:\n%s",
			hardRequireNoGrokHomeUUID, resp.Elapsed, resp.Stderr)
	} else if id != hardRequireNoGrokHomeUUID {
		t.Fatalf("runner_session_id=%q want %q", id, hardRequireNoGrokHomeUUID)
	}
}
```
