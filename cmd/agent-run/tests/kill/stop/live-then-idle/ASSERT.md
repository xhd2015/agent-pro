## Expected

- Exit code 0.
- Stdout is exactly `stopped kill-live-1\n`.
- Fixture process is no longer alive.
- Registry entry is removed.

## Side Effects

- Keep-alive / serve PID terminated; registry free for re-use.

## Exit Code

0

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	sid := req.SessionID
	if sid == "" {
		sid = "kill-live-1"
	}
	assertStdoutLine(t, resp.Stdout, "stopped "+sid)

	pid := fixturePID(t, req)
	// Reclaim uses TERM then brief wait then KILL — allow a short settle.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && processAlive(pid) {
		time.Sleep(50 * time.Millisecond)
	}
	if processAlive(pid) {
		t.Fatalf("expected fixture pid %d dead after kill", pid)
	}
	if registryExists(req.Home, sid) {
		t.Fatalf("expected registry entry removed for %s", sid)
	}
}
```
