## Expected

- No API error.
- `resp.Args` (launch argv) exactly equals `resp.ExpectedArgs` (`BuildArgs` of
  filled detach `RunOpts`).
- Expected fill includes detach profile + dir + allow-relocate + grok-tty.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	if len(resp.ExpectedArgs) == 0 {
		t.Fatal("expected ExpectedArgs from harness fill")
	}
	assertArgsEqual(t, resp.Args, resp.ExpectedArgs)
	// Sanity: detach fill mapping locked
	assertHasArg(t, resp.ExpectedArgs, "--agent-runner=grok-tty")
	assertHasArg(t, resp.ExpectedArgs, "--dir=/compose/detach-ws")
	assertHasArg(t, resp.ExpectedArgs, "--allow-relocate-resume-session-dir")
	assertHasArg(t, resp.ExpectedArgs, "--detach")
	assertNotHasArg(t, resp.ExpectedArgs, "--open")
	assertNotHasArg(t, resp.ExpectedArgs, "--new-terminal")
}
```
