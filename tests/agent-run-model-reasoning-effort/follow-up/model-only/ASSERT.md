## Expected

- No API error.
- Model flag present with value `o3`.
- No `--model-reasoning-effort` flag (empty effort must not invent `max`).

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
	assertNoEffortFlag(t, fu)
	assertContains(t, fu, "sess-fu-model-only")
	assertNotContains(t, fu, "--new-terminal")
}
```
