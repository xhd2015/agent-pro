## Expected

- Exit code 0.
- No `banner not detected` / `TUI banner not detected` on stdout or stderr.
- Exactly one `grok-tty: <session-id>` line on stderr.
- Stdout must not contain `grok-tty:`.
- No forbidden open discovery/event noise.

## Errors

- Must not fail solely because ready markers never appeared within the old
  hard-timeout window.

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
	combined := resp.Stdout + "\n" + resp.Stderr
	if hasBannerNotDetected(combined) {
		t.Fatalf("--open + prompt must not hard-fail on missing banner/OpenReady:\n%s", resp.Stderr)
	}
	assertSuccess(t, resp)

	id, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty")
	if !ok {
		t.Fatalf("missing post-attach grok-tty session id on stderr:\n%s", resp.Stderr)
	}
	if n := countPrefixedSessionIDLines(resp.Stderr, "grok-tty"); n != 1 {
		t.Fatalf("want exactly 1 grok-tty session id line, got %d (id=%q)\nstderr:\n%s", n, id, resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "grok-tty:") {
		t.Fatalf("session id must not appear on stdout:\n%s", resp.Stdout)
	}
	if noise := forbiddenOpenNoise(combined); len(noise) > 0 {
		t.Fatalf("unexpected open noise %v\nstdout:\n%s\nstderr:\n%s", noise, resp.Stdout, resp.Stderr)
	}
}
```
