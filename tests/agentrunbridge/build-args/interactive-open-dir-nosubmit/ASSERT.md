## Expected

Exact argv includes `--dir=/tmp/ws-bridge` and `--no-submit` before `--open`:

```text
run
--session-id=sess-open-dir
--agent-runner=grok-tty
--auto-send-or-resume
--new-terminal
--dir=/tmp/ws-bridge
--no-submit
--open
--
with dir
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
		"--session-id=sess-open-dir",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--new-terminal",
		"--dir=/tmp/ws-bridge",
		"--no-submit",
		"--open",
		"--",
		"with dir",
	}
	assertArgsEqual(t, resp.Args, want)
}
```
