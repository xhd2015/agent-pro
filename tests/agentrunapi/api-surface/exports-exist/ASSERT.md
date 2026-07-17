## Expected

- `Classify` on missing session returns `ModeRun` and `found=false`.
- No harness error (package must exist and export symbols).
- `AutoSendOrResume` is callable (empty session may set API error; allowed).

## Side Effects

- None beyond temp store home mkdir.

## Errors

- No harness error. API validation error on empty auto session is OK.

## Exit Code

N/A (package call)

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Mode", resp.Mode, agentrunapi.ModeRun)
	assertEqual(t, "Found", resp.Found, false)
}
```
