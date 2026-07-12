## Expected

- Argv has no `--agent-runner` and no `--agent-runner=` prefix.
- Session id and prompt still present.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertNotHasArgPrefix(t, resp.Args, "--agent-runner")
	assertHasArg(t, resp.Args, "--session-id=sess-no-runner")
	if len(resp.Args) == 0 || resp.Args[len(resp.Args)-1] != "no runner flag" {
		t.Fatalf("expected prompt last, got %q", resp.Args)
	}
}
```
