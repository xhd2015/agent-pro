## Expected

- No API error.
- Line contains custom binary path and each prefix token.
- Contains `run` and `--auto-send-or-resume`.
- Does **not** contain `--new-terminal`.
- Does not use bare default `agent-run` as the **driver** sole binary when custom path is set
  (prefix may still contain the string `agent-exec`; binary path must appear).

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
	assertContains(t, resp.FollowUp, "/usr/local/bin/spl-helper")
	assertContains(t, resp.FollowUp, "local-bot")
	assertContains(t, resp.FollowUp, "agent-exec")
	assertContains(t, resp.FollowUp, "run")
	assertContains(t, resp.FollowUp, "--auto-send-or-resume")
	assertNotContains(t, resp.FollowUp, "--new-terminal")
}
```
