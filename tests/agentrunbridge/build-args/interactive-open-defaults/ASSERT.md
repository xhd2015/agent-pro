## Expected

Exact argv (SeaTalk `buildAgentRunAutoOpenArgs` spirit):

```text
run
--session-id=sess-open-1
--agent-runner=grok-tty
--auto-send-or-resume
--new-terminal
--open
--
open me
```

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
		"--session-id=sess-open-1",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--new-terminal",
		"--open",
		"--",
		"open me",
	}
	assertArgsEqual(t, resp.Args, want)
}
```
