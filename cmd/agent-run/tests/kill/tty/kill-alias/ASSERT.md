## Expected

- Exit code 0.
- Stdout is exactly `stopped kill-tty-alias-1\n`.
- Fixture process dead; registry removed (same as top-level kill).

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
		sid = "kill-tty-alias-1"
	}
	assertStdoutLine(t, resp.Stdout, "stopped "+sid)

	pid := fixturePID(t, req)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && processAlive(pid) {
		time.Sleep(50 * time.Millisecond)
	}
	if processAlive(pid) {
		t.Fatalf("expected fixture pid %d dead after tty kill", pid)
	}
	if registryExists(req.Home, sid) {
		t.Fatalf("expected registry entry removed for %s", sid)
	}
}
```
