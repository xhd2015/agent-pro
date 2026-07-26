---
label: e2e
---

## Expected Output

Stdout includes both labeled id lines (order: session-id then terminal-id preferred).

```
---
version: 2
__SESSION_ID__: type=string, example=sess-1, agent storage session id
__TERMINAL_ID__: type=string, example=term-1, tty registry session id
---
session-id: __SESSION_ID__
terminal-id: __TERMINAL_ID__
```

(Assert uses labeled parse rather than strict full-match if optional blank lines
appear; both labels and non-empty token values are required.)

## Expected

- Exit code 0.
- `session-id:` line present on **stdout**.
- `terminal-id:` line present on **stdout**.
- Ids are non-empty single tokens (no ANSI).

## Exit Code

0

```go
import (
	"regexp"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)
	sid, tid := assertDetachIDsOnStdout(t, resp)

	// No ANSI CSI sequences in the id lines / stdout contract.
	if strings.Contains(resp.Stdout, "\x1b[") {
		t.Fatalf("detach stdout must not contain ANSI escapes:\n%q", resp.Stdout)
	}
	idRe := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
	if !idRe.MatchString(sid) || !idRe.MatchString(tid) {
		t.Fatalf("ids must be single tokens; session=%q terminal=%q", sid, tid)
	}
}
```
