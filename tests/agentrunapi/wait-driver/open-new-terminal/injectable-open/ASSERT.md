## Expected

- No API error.
- `OpenCalls == 1`.
- `OpenDir` is workspace dir.
- `OpenFollowUp` contains `agent-run`, `--auto-send-or-resume`, `--open`, session id;
  no `--new-terminal`.

## Side Effects

- Only the injectable OpenTerminal hook (no real iTerm / osascript).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "OpenCalls", resp.OpenCalls, 1)
	assertEqual(t, "OpenDir", resp.OpenDir, "/tmp/ws-iterm")
	assertContains(t, resp.OpenFollowUp, "agent-run")
	assertContains(t, resp.OpenFollowUp, "--auto-send-or-resume")
	assertContains(t, resp.OpenFollowUp, "--open")
	assertContains(t, resp.OpenFollowUp, "sess-iterm-1")
	assertNotContains(t, resp.OpenFollowUp, "--new-terminal")
}
```
