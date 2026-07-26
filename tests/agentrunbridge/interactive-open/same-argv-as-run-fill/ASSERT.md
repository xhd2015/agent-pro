## Expected

- No API error.
- `resp.Args` (launch argv) exactly equals `resp.ExpectedArgs` (`BuildArgs` of filled `RunOpts`).
- Expected fill includes open profile + dir + nosubmit + grok-tty.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	if len(resp.ExpectedArgs) == 0 {
		t.Fatal("expected ExpectedArgs from harness fill")
	}
	assertArgsEqual(t, resp.Args, resp.ExpectedArgs)
	// Sanity: fill mapping locked
	assertHasArg(t, resp.ExpectedArgs, "--agent-runner=grok-tty")
	assertHasArg(t, resp.ExpectedArgs, "--dir=/compose/ws")
	assertHasArg(t, resp.ExpectedArgs, "--no-submit")
	assertHasArg(t, resp.ExpectedArgs, "--open")
}
```
