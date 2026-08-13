## Expected

- Follow-up uses the alt session id (from `--pid 7000`), not the default 4242 session.

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
	if !strings.Contains(call.FollowUp, "--session-id "+fixtureAltSessionID) {
		t.Fatalf("follow-up should use --pid session %s: %q", fixtureAltSessionID, call.FollowUp)
	}
	if strings.Contains(call.FollowUp, fixtureSessionID) {
		t.Fatalf("follow-up used default start-pid session, --pid ignored: %q", call.FollowUp)
	}
	assertStdoutExact(t, resp.Stdout, modeASuccessLine(fixtureAltSessionID))
}
```
