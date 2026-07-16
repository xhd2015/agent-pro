## Expected Output

Stderr ends with a single prefixed registry session id (no pre-attach chatter).

```
---
version: 3
__ID__: type=string, example=session-1, terminal session id
---
grok-tty: __ID__
```

(Actual stderr may include only that line, or implementers may print solely that
line after attach; assert uses token parse + count rather than strict full-match
if extra blank noise appears — prefer single id line.)

## Expected

- Exit code 0.
- Exactly one `grok-tty: <id>` session-id line on stderr.
- Stdout does not contain `grok-tty:`.
- Combined output has no forbidden open noise.

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
	if noise := forbiddenOpenNoise(resp.Stdout + "\n" + resp.Stderr); len(noise) > 0 {
		t.Fatalf("unexpected open noise %v\nstdout:\n%s\nstderr:\n%s", noise, resp.Stdout, resp.Stderr)
	}
}
```
