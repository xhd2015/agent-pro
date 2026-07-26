## Expected

- No harness panic/error; API error empty.
- Session found (`found=true`).
- Mode is `run` (LifecycleProbe default without live TTY → else branch).
- MetaID matches SessionID.

## Side Effects

- LifecycleProbe may touch store home paths; must not require real iTerm/TCP.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "Mode", resp.Mode, agentrunapi.ModeRun)
	assertEqual(t, "Found", resp.Found, true)
	assertEqual(t, "MetaID", resp.MetaID, req.SessionID)
}
```
