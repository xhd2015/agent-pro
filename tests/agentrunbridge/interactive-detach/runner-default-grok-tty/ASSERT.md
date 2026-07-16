## Expected

- Launch argv contains `--agent-runner=grok-tty`.
- Detach profile present (`--detach`, no `--open`).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertHasArg(t, resp.Args, "--agent-runner=grok-tty")
	assertHasArg(t, resp.Args, "--detach")
	assertNotHasArg(t, resp.Args, "--open")
	assertNotHasArg(t, resp.Args, "--new-terminal")
}
```
