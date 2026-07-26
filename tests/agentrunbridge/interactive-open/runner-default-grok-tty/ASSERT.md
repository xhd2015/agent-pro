## Expected

- Launch argv includes `--agent-runner=grok-tty`.
- No API error.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertHasArg(t, resp.Args, "--agent-runner=grok-tty")
}
```
