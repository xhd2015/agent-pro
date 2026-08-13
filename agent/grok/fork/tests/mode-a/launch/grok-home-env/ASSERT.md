## Expected

- Follow-up prefixes `GROK_HOME=<home>` (shell-quoted if needed) before the executable.

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
	homeAssign := "GROK_HOME=" + shell.ShellQuote(req.GrokHome)
	alt := "GROK_HOME=" + req.GrokHome
	if !strings.HasPrefix(call.FollowUp, homeAssign) && !strings.HasPrefix(call.FollowUp, alt) {
		t.Fatalf("follow-up must prefix GROK_HOME=…: %q", call.FollowUp)
	}
	if !strings.Contains(call.FollowUp, "--session-id "+fixtureSessionID) {
		t.Fatalf("follow-up missing --session-id: %q", call.FollowUp)
	}
}
```
