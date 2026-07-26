## Expected

- No API error.
- `SendCalls == 1`, `RunCalls == 0`, `ResumeCalls == 0`.
- Observed mode is `send`.

## Side Effects

- Only SendLive hook (no agentsend / no agent-run binary).

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
	assertEqual(t, "RunCalls", resp.RunCalls, 0)
	assertEqual(t, "SendCalls", resp.SendCalls, 1)
	assertEqual(t, "ResumeCalls", resp.ResumeCalls, 0)
}
```
