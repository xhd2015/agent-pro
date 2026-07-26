## Expected

- `Mode` is `send`.
- `found` is true.
- Meta session id matches request.
- No API error.

## Side Effects

- Session meta remains as seeded (Classify is read-only).

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
	assertEqual(t, "Mode", resp.Mode, agentrunapi.ModeSend)
	assertEqual(t, "Found", resp.Found, true)
	assertEqual(t, "MetaID", resp.MetaID, req.SessionID)
}
```
