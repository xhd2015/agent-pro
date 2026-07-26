## Expected

- Argv contains `run`, `--session-id=sess-keep`, `--keep-tty`, `--agent-runner=fake-opencode`, and the prompt.
- Argv does **not** contain `--open`.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertHasArg(t, resp.Args, "run")
	assertHasArg(t, resp.Args, "--session-id=sess-keep")
	assertHasArg(t, resp.Args, "--keep-tty")
	assertHasArg(t, resp.Args, "--agent-runner=fake-opencode")
	assertNotHasArg(t, resp.Args, "--open")
	// Prompt is final arg (after optional -- separator if implementer uses it for keep-tty).
	if len(resp.Args) == 0 || resp.Args[len(resp.Args)-1] != "keep tty prompt" {
		t.Fatalf("expected prompt as last arg, got %q", resp.Args)
	}
}
```
