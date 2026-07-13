## Expected

Argv includes `--allow-relocate-resume-session-dir` (with open profile + dir):

```text
run
--session-id=sess-relocate-1
--agent-runner=grok-tty
--auto-send-or-resume
--new-terminal
--dir=/tmp/ws-new
--allow-relocate-resume-session-dir
--open
--
relocate ok
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
		"--session-id=sess-relocate-1",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--new-terminal",
		"--dir=/tmp/ws-new",
		"--allow-relocate-resume-session-dir",
		"--open",
		"--",
		"relocate ok",
	}
	assertArgsEqual(t, resp.Args, want)
}
```
