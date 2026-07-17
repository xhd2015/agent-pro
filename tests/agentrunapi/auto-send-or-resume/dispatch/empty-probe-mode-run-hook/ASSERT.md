## Expected

- No API error.
- `RunCalls == 1`, `SendCalls == 0`, `ResumeCalls == 0`.
- Observed mode is `run`.

## Side Effects

- Only RunSession hook; EmptyProbe avoids TTY I/O.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "Mode", resp.Mode, agentrunapi.ModeRun)
	assertEqual(t, "RunCalls", resp.RunCalls, 1)
	assertEqual(t, "SendCalls", resp.SendCalls, 0)
	assertEqual(t, "ResumeCalls", resp.ResumeCalls, 0)
}
```
