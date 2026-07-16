## Expected

Launch argv includes dir, allow-relocate, detach; no open/new-terminal:

```text
run
--session-id=sess-id-dir
--agent-runner=grok-tty
--auto-send-or-resume
--dir=/tmp/ws-id
--allow-relocate-resume-session-dir
--detach
--
detach dir
```

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	want := []string{
		"run",
		"--session-id=sess-id-dir",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--dir=/tmp/ws-id",
		"--allow-relocate-resume-session-dir",
		"--detach",
		"--",
		"detach dir",
	}
	assertArgsEqual(t, resp.Args, want)
	assertNotHasArg(t, resp.Args, "--open")
	assertNotHasArg(t, resp.Args, "--new-terminal")
}
```
