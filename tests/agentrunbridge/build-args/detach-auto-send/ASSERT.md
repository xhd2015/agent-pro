## Expected

Exact argv (detach profile; SeaTalk parity prompt after `--`):

```text
run
--session-id=sess-detach-1
--agent-runner=grok-tty
--auto-send-or-resume
--detach
--
detach me
```

Must not include `--open` or `--new-terminal`.

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A (package call)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	want := []string{
		"run",
		"--session-id=sess-detach-1",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--detach",
		"--",
		"detach me",
	}
	assertArgsEqual(t, resp.Args, want)
	assertNotHasArg(t, resp.Args, "--open")
	assertNotHasArg(t, resp.Args, "--new-terminal")
}
```
