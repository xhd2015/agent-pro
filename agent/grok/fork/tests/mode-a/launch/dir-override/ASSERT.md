## Expected

- Open dir is the `--dir` override, not session cwd.
- Follow-up contains `--session-id` and `--dir <override>`.

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
	if call.Dir != req.OverrideDir {
		t.Fatalf("open dir: got %q, want override %q", call.Dir, req.OverrideDir)
	}
	if !strings.Contains(call.FollowUp, "--session-id "+fixtureSessionID) {
		t.Fatalf("follow-up missing --session-id: %q", call.FollowUp)
	}
	if !strings.Contains(call.FollowUp, "--dir "+shell.ShellQuote(req.OverrideDir)) &&
		!strings.Contains(call.FollowUp, "--dir "+req.OverrideDir) {
		t.Fatalf("follow-up missing --dir override: %q", call.FollowUp)
	}
}
```
