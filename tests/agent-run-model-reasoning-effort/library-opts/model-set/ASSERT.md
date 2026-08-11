## Expected

- No API error.
- RunSession called once.
- CapturedModel == `o3`.
- CapturedEffort remains empty (not invent `max`).

## Side Effects

- Temp store home under `t.TempDir()` only.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "RunCalls", resp.RunCalls, 1)
	assertEqual(t, "CapturedModel", resp.CapturedModel, fixtureModel)
	assertEqual(t, "CapturedEffort", resp.CapturedEffort, "")
}
```
