## Expected

Exact argv includes open profile plus `--color` **after** `--open` and
**before** `--` / prompt:

```text
run
--session-id=sess-color-1
--agent-runner=grok-tty
--auto-send-or-resume
--new-terminal
--open
--color
--
with color
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
		"--session-id=sess-color-1",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--new-terminal",
		"--open",
		"--color",
		"--",
		"with color",
	}
	assertArgsEqual(t, resp.Args, want)
}
```
