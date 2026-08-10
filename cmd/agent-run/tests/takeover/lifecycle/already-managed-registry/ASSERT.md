## Expected

- Exit code 0.
- Stderr contains a `warning:` that the session is already managed by agent-run
  (nothing to take over).
- Fixture sleep PID still alive.
- No kill log entries.
- No iTerm ForceNew script written.

## Side Effects

- No kill of registry PID; no second import.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	assertExitCode(t, resp, 0)
	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "warning") {
		t.Fatalf("expected warning: on stderr/stdout for already-managed, got:\n%s", combined)
	}
	assertContainsAny(t, combined,
		"already managed",
		"already under agent-run",
		"nothing to take over",
		"managed by agent-run",
	)
	pid := fixturePID(t, req)
	if !processAlive(pid) {
		t.Fatalf("already-managed must not kill registry fixture pid %d", pid)
	}
	assertNoKillLog(t, req)
	assertNoItermScript(t, req)
}
```
