## Expected

- No error.
- Exactly 2 user prompts remain (matching `alpha` case-insensitively).
- Texts contain `alpha one` and `prefix ALPHA suffix` in chrono order.
- Non-matching `beta noise only` is absent.
- OmittedBefore and OmittedAfter are 0 (no head/tail).

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	sp := resp.Single
	if sp == nil {
		t.Fatal("Single is nil")
	}
	assertPromptCount(t, sp, 2)
	assertPromptText(t, sp, 0, "alpha one")
	if !strings.Contains(sp.UserPrompts[1].Text, "ALPHA") {
		t.Fatalf("second kept prompt want ALPHA match, got %q", sp.UserPrompts[1].Text)
	}
	for _, p := range sp.UserPrompts {
		if strings.Contains(p.Text, "beta noise") {
			t.Fatalf("non-matching prompt not filtered: %q", p.Text)
		}
	}
	if sp.OmittedBefore != 0 || sp.OmittedAfter != 0 {
		t.Fatalf("OmittedBefore=%d OmittedAfter=%d want 0,0", sp.OmittedBefore, sp.OmittedAfter)
	}
}
```
