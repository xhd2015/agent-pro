## Expected

- No API error.
- `ResumeCalls == 1`, `RunCalls == 0`, `SendCalls == 0`.
- Observed mode is `resume`.

## Side Effects

- Only ResumeSession hook (no provider spawn / no agent-run binary).

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
	assertEqual(t, "Mode", resp.Mode, agentrunapi.ModeResume)
	assertEqual(t, "RunCalls", resp.RunCalls, 0)
	assertEqual(t, "SendCalls", resp.SendCalls, 0)
	assertEqual(t, "ResumeCalls", resp.ResumeCalls, 1)
}
```
