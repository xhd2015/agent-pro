## Expected

Exact launch argv:

```text
run
--session-id=sess-io-dir
--agent-runner=grok-tty
--auto-send-or-resume
--new-terminal
--dir=/work/bridge
--no-submit
--open
--
dir prompt
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
		"--session-id=sess-io-dir",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--new-terminal",
		"--dir=/work/bridge",
		"--no-submit",
		"--open",
		"--",
		"dir prompt",
	}
	assertArgsEqual(t, resp.Args, want)
}
```
