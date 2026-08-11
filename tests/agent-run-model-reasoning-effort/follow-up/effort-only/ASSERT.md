## Expected

- No API error.
- Reasoning-effort flag present with value `max`.
- No `--model` flag (empty model must not invent `gpt-5.6-luna`).

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
	assertHasEffortValue(t, fu, fixtureEffortMax)
	assertNoModelFlag(t, fu)
	// Prefix-safe: hasReasoningEffortFlag true does not count as model.
	if hasModelFlag(fu) {
		t.Fatalf("empty model must not emit --model; line=%q", fu)
	}
	assertContains(t, fu, "sess-fu-effort-only")
	assertNotContains(t, fu, "--new-terminal")
}
```
