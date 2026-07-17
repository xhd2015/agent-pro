## Expected

- No API error.
- Contains session id, `--auto-send-or-resume`, `--detach`, dir, allow-relocate flag, prompt.
- No `--open`, no `--new-terminal`.

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
	assertContains(t, fu, "sess-detach-shape")
	assertContains(t, fu, "--auto-send-or-resume")
	assertContains(t, fu, "--detach")
	assertContains(t, fu, "/tmp/ws-detach")
	assertContains(t, fu, "--allow-relocate-resume-session-dir")
	assertContains(t, fu, "detach child")
	assertNotContains(t, fu, "--open")
	assertNotContains(t, fu, "--new-terminal")
}
```
