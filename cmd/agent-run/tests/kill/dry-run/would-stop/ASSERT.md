## Expected

- Exit code 0.
- Stdout is exactly `dry-run: would stop kill-dry-1\n`.
- Fixture process still alive after dry-run.
- Registry file still present under `grok-tty-registry/`.

## Side Effects

- No process termination; no registry removal.

## Exit Code

0

```go
import (
	"testing"

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
		sid = "kill-dry-1"
	}
	assertStdoutLine(t, resp.Stdout, "dry-run: would stop "+sid)

	pid := fixturePID(t, req)
	if !processAlive(pid) {
		t.Fatalf("dry-run must not terminate fixture pid %d", pid)
	}
	if !registryExists(req.Home, sid) {
		t.Fatalf("dry-run must not remove registry for %s", sid)
	}
}
```
