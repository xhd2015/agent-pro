## Expected

- Argv includes `--auto-send-or-resume`.
- Argv does not include `--open`.
- Session id and prompt present.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertHasArg(t, resp.Args, "--auto-send-or-resume")
	assertNotHasArg(t, resp.Args, "--open")
	assertHasArg(t, resp.Args, "--session-id=sess-autosend")
	if len(resp.Args) == 0 || resp.Args[len(resp.Args)-1] != "auto only" {
		t.Fatalf("expected prompt last, got %q", resp.Args)
	}
}
```
