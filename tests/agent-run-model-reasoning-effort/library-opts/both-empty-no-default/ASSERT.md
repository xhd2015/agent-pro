## Expected

- No API error.
- RunSession called once.
- CapturedModel == `""` and CapturedEffort == `""`.
- Must not invent `gpt-5.6-luna` or `max` as defaults.

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
	assertEqual(t, "CapturedModel", resp.CapturedModel, "")
	assertEqual(t, "CapturedEffort", resp.CapturedEffort, "")
	if resp.CapturedModel == forbiddenDefaultModel {
		t.Fatalf("must not invent default model %q", forbiddenDefaultModel)
	}
	if resp.CapturedEffort == fixtureEffortMax {
		t.Fatalf("must not invent default effort %q when empty", fixtureEffortMax)
	}
}
```
