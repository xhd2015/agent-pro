## Expected

- No API error.
- Neither `--model` nor `--model-reasoning-effort` present.
- Line does not invent forbidden default model token as a model flag value.
- Open profile still works; no `--new-terminal`.

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
	assertNoModelFlag(t, fu)
	assertNoEffortFlag(t, fu)
	// Extra guard: if a future default sneaks in as --model=gpt-5.6-luna
	if v, ok := modelFlagValue(fu); ok && v == forbiddenDefaultModel {
		t.Fatalf("must not invent default model %q; line=%q", forbiddenDefaultModel, fu)
	}
	assertContains(t, fu, "sess-fu-both-empty")
	assertContains(t, fu, "--auto-send-or-resume")
	assertContains(t, fu, "--open")
	assertNotContains(t, fu, "--new-terminal")
}
```
