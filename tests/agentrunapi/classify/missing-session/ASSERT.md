## Expected

- `Mode` is `run` (`agentrunapi.ModeRun`).
- `found` is false.
- No API error.

## Side Effects

- No session directory created by Classify.

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
	assertEqual(t, "Found", resp.Found, false)
}
```
