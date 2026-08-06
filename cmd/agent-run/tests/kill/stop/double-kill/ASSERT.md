## Expected

- Exit code 0 (idempotent).
- Stderr contains `warning:` and indicates the session is not running.
- Stderr mentions the session id.

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
	assertExitCode(t, resp, 0)
	sid := req.SessionID
	if sid == "" {
		sid = "kill-double-1"
	}
	stderr := resp.Stderr
	lower := strings.ToLower(stderr)
	if !strings.Contains(lower, "warning:") && !strings.Contains(lower, "warning") {
		t.Fatalf("expected warning on stderr for double kill, got:\n%s", stderr)
	}
	if !strings.Contains(lower, "not running") {
		t.Fatalf("expected 'not running' on stderr for double kill, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, sid) {
		t.Fatalf("expected session id %q in warning stderr:\n%s", sid, stderr)
	}
}
```
