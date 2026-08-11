## Expected

- No API error.
- RunSession called once.
- CapturedEffort == `max`.
- CapturedModel remains empty (not invent luna).

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
	assertEqual(t, "CapturedEffort", resp.CapturedEffort, fixtureEffortMax)
	assertEqual(t, "CapturedModel", resp.CapturedModel, "")
	if resp.CapturedModel == forbiddenDefaultModel {
		t.Fatalf("must not invent default model %q", forbiddenDefaultModel)
	}
}
```
