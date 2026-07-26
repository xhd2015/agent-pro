## Expected

- No API error.
- Follow-up line contains `agent-run` (default driver; may be shell-quoted).
- Contains `run` and `--auto-send-or-resume`.
- Does **not** contain `--new-terminal`.

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertContains(t, resp.FollowUp, "agent-run")
	assertContains(t, resp.FollowUp, "run")
	assertContains(t, resp.FollowUp, "--auto-send-or-resume")
	assertNotContains(t, resp.FollowUp, "--new-terminal")
}
```
