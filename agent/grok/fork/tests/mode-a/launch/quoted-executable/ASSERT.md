## Expected

- Follow-up equals `ShellQuote(executable) + " --session-id " + id`.
- Unquoted path-with-spaces must not appear as a raw token.

## Side Effects

- One recorded open.

## Errors

- None.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/shell"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	call := assertOneOpen(t, resp)
	quoted := shell.ShellQuote(req.Executable)
	want := quoted + " --session-id " + fixtureSessionID
	if call.FollowUp != want {
		t.Fatalf("follow-up: got %q, want %q", call.FollowUp, want)
	}
	if quoted == req.Executable {
		t.Fatal("test fixture path should require quoting (contains space)")
	}
	if strings.Contains(call.FollowUp, req.Executable+" --session-id") {
		t.Fatalf("executable with spaces was not quoted: %q", call.FollowUp)
	}
}
```
