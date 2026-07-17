## Expected

- No API error.
- Contains `--session-id=` (or session id token), `--agent-runner=`, `--auto-send-or-resume`,
  `--dir=`, `--no-submit`, `--open`, prompt text.
- Contains `--` separator before prompt (open profile).
- No `--new-terminal`, no `--detach`.

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	fu := resp.FollowUp
	assertContains(t, fu, "sess-open-shape")
	assertContains(t, fu, "grok-tty")
	assertContains(t, fu, "--auto-send-or-resume")
	assertContains(t, fu, "/tmp/ws-open")
	assertContains(t, fu, "--no-submit")
	assertContains(t, fu, "--open")
	assertContains(t, fu, "open child")
	// open profile: -- separator before prompt
	assertContains(t, fu, "--")
	assertNotContains(t, fu, "--new-terminal")
	assertNotContains(t, fu, "--detach")
}
```
