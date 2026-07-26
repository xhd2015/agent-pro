## Expected

Exact argv includes `--dir=/tmp/ws-detach` and
`--allow-relocate-resume-session-dir` before `--detach`:

```text
run
--session-id=sess-detach-dir
--agent-runner=grok-tty
--auto-send-or-resume
--dir=/tmp/ws-detach
--allow-relocate-resume-session-dir
--detach
--
detach with dir
```

Must not include `--open` or `--new-terminal`.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	want := []string{
		"run",
		"--session-id=sess-detach-dir",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--dir=/tmp/ws-detach",
		"--allow-relocate-resume-session-dir",
		"--detach",
		"--",
		"detach with dir",
	}
	assertArgsEqual(t, resp.Args, want)
	assertNotHasArg(t, resp.Args, "--open")
	assertNotHasArg(t, resp.Args, "--new-terminal")
}
```
