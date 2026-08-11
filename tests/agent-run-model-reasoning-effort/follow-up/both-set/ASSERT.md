## Expected

- No API error.
- Follow-up contains model flag with value `o3`.
- Follow-up contains reasoning-effort flag with value `high`.
- Still has open/auto-send profile; no `--new-terminal`.

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
	fu := resp.FollowUp
	assertHasModelValue(t, fu, fixtureModel)
	assertHasEffortValue(t, fu, fixtureEffort)
	assertContains(t, fu, "sess-fu-both")
	assertContains(t, fu, "--auto-send-or-resume")
	assertContains(t, fu, "--open")
	assertNotContains(t, fu, "--new-terminal")
}
```
