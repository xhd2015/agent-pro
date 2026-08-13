## Expected

- Follow-up / success line use the nearest session id, not the main grok’s id.

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
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	call := assertOneOpen(t, resp)
	if !strings.Contains(call.FollowUp, "--session-id "+fixtureSessionID) {
		t.Fatalf("follow-up should use nearest session %s: %q", fixtureSessionID, call.FollowUp)
	}
	if strings.Contains(call.FollowUp, fixtureMainSessionID) {
		t.Fatalf("follow-up used topmost main session: %q", call.FollowUp)
	}
	assertStdoutExact(t, resp.Stdout, modeASuccessLine(fixtureSessionID))
}
```
