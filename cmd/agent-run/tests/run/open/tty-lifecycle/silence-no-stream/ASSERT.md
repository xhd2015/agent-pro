## Expected

- Exit code 0.
- Combined stdout+stderr must **not** contain discovery/event stream noise:
  `Resolve session id`, `💭`, `💬`, `[done]`, or NDJSON type markers.
- Allowed: optional final `grok-tty: <id>` line on stderr after attach (asserted
  strictly in the sibling leaf). Stdout should not stream formatted agent events.

## Side Effects

- None required beyond successful open run.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	combined := resp.Stdout + "\n" + resp.Stderr
	if noise := forbiddenOpenNoise(combined); len(noise) > 0 {
		t.Fatalf("--open must be silent (no discovery/event stream); found %v\nstdout:\n%s\nstderr:\n%s",
			noise, resp.Stdout, resp.Stderr)
	}

	// Session id prefix must not appear on stdout (stderr-only, post-attach).
	if strings.Contains(resp.Stdout, "grok-tty:") {
		t.Fatalf("session id / runner prefix must not appear on stdout under --open:\n%s", resp.Stdout)
	}
}
```
