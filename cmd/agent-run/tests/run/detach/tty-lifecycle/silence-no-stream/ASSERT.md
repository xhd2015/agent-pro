---
label: e2e
---

## Expected

- Exit code 0.
- Combined stdout+stderr must **not** contain discovery/event stream noise:
  `Resolve session id`, `💭`, `💬`, `[done]`, or NDJSON type markers.
- Allowed: `session-id:` / `terminal-id:` on stdout; optional soft-bind
  `grok-tty: grok session …` lines on stderr.

## Side Effects

- None required beyond successful detach run.

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

	combined := resp.Stdout + "\n" + resp.Stderr
	if noise := forbiddenDetachNoise(combined); len(noise) > 0 {
		t.Fatalf("--detach must be silent (no discovery/event stream); found %v\nstdout:\n%s\nstderr:\n%s",
			noise, resp.Stdout, resp.Stderr)
	}
	assertDetachIDsOnStdout(t, resp)
}
```
