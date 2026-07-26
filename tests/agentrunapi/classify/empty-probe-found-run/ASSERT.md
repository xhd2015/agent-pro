## Expected

- `Mode` is `run` even though session was found.
- `found` is true.
- No API error.
- Uses `EmptyProbe` (unknown lifecycle), not live/resume.

## Side Effects

- Classify is read-only; EmptyProbe does no TTY I/O.

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
