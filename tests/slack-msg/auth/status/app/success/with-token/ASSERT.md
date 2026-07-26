---
label: unit
explanation: "app status via apps.connections.open; note line; masked xapp token"
---

## Expected Output

```
Using config from: (none)
kind: app
ok: true
token: xapp-...oken
note: app-level token (Socket Mode / connections); not used for channels/send/history
```

## Expected

- Exit code 0.
- Exact human lines including fixed `note:` (trailing newline).
- Validation uses `apps.connections.open` (implementer).
- Stdout must not contain full `xapp-slacktest-token`.
- Stderr empty.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, slackTestAppToken) {
		t.Fatalf("stdout must not contain raw app token %q:\n%s", slackTestAppToken, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
Using config from: (none)
kind: app
ok: true
token: xapp-...oken
note: app-level token (Socket Mode / connections); not used for channels/send/history
`)
}
```
