---
label: e2e
---

## Expected

- Exit code 0.
- `StreamProbeSeen` is true: stdout contains `STREAM_TURN_TWO_MARKER` **before** the fake
  TUI exits (while the PTY session is still running).
- Turn 2 marker appears after turn 1 completed — tail did not stop at first `turn_completed`.
- Stdout also shows turn 1 user prompt text from pre-seeded bootstrap.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if !resp.StreamProbeSeen {
		t.Fatalf("expected turn 2 marker %q on stdout before timeout; stdout:\n%s\nstderr:\n%s",
			streamTurnTwoMarker, resp.Stdout, resp.Stderr)
	}
	if !resp.StreamProbeBeforeExit {
		t.Fatalf("expected turn 2 marker while PTY still running (before fake TUI exit); stdout:\n%s", resp.Stdout)
	}
	stdout := strings.ToLower(resp.Stdout)
	if !strings.Contains(stdout, strings.ToLower(streamTurnTwoMarker)) {
		t.Fatalf("stdout missing turn 2 marker; stdout:\n%s", resp.Stdout)
	}
	if !strings.Contains(stdout, strings.ToLower(turnOnePromptText)) {
		t.Fatalf("stdout missing turn 1 user prompt from bootstrap; stdout:\n%s", resp.Stdout)
	}
}
```