## Expected

- Process wall time ≥ materialization delay (proves open waited past attach for
  delayed discovery; does not soft-exit unbound early).
- Exit code 0.
- Stderr contains `grok-tty: grok session` with the delayed UUID.
- Stderr contains `grok-tty: grok updates` / `updates.jsonl`.
- Meta under home has `runner_session_id` equal to the delayed UUID.

## Side Effects

- Delayed `GROK_HOME` session dir is written only after the configured delay;
  bind worker (or post-detach join) must still observe it.

## Exit Code

0

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v elapsed=%s\nstdout:\n%s\nstderr:\n%s", err, resp.Elapsed, resp.Stdout, resp.Stderr)
	}
	// Critical: must have waited at least until delayed materialization.
	// Allow small timer slack but require clear wait past soft 750ms skip.
	minWait := bgBindMaterializeDelay - 200*time.Millisecond
	if minWait < time.Second {
		minWait = time.Second
	}
	if resp.Elapsed < minWait {
		t.Fatalf("open exited too quickly (elapsed=%s < minWait=%s); expected wait for delayed bind\nstderr:\n%s",
			resp.Elapsed, minWait, resp.Stderr)
	}
	assertSuccess(t, resp)
	assertContains(t, resp.Stderr, "grok-tty: grok session "+bgBindDelayedUUID)
	assertContains(t, resp.Stderr, "grok-tty: grok updates")
	assertContains(t, resp.Stderr, "updates.jsonl")
	if _, id, ok := findMetaRunnerSessionID(t, req.Home, "grok-tty", bgBindDelayedUUID); !ok {
		t.Fatalf("no meta.json with runner_session_id=%q after delayed bind wait (elapsed=%s)\nstderr:\n%s",
			bgBindDelayedUUID, resp.Elapsed, resp.Stderr)
	} else if id != bgBindDelayedUUID {
		t.Fatalf("runner_session_id=%q want %q", id, bgBindDelayedUUID)
	}
}
```
