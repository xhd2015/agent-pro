## Expected

- Exactly 1 prompt kept: text contains both `fix` and `timeout`.
- Partial matches (`fix` only / `timeout` only) are dropped.
- No head/tail omissions.

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
	assertPromptCount(t, sp, 1)
	assertPromptText(t, sp, 0, "please fix the timeout path")
	for _, p := range sp.UserPrompts {
		if !strings.Contains(strings.ToLower(p.Text), "fix") ||
			!strings.Contains(strings.ToLower(p.Text), "timeout") {
			t.Fatalf("kept prompt missing AND tokens: %q", p.Text)
		}
	}
	if sp.OmittedBefore != 0 || sp.OmittedAfter != 0 {
		t.Fatalf("OmittedBefore=%d OmittedAfter=%d want 0,0", sp.OmittedBefore, sp.OmittedAfter)
	}
}
```
