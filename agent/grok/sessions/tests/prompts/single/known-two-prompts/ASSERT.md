## Expected

- No error.
- Exactly 2 user prompts in chronological order.
- Texts: `hello world`, `second prompt`.
- Timestamps match wire: 14:30:00Z and 14:45:00Z (fixedNow − 30m / − 15m).
- Session.ID equals fixture id.

## Errors

- None.

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	sp := resp.Single
	if sp == nil {
		t.Fatal("Single is nil")
	}
	if sp.ID != req.SessionID {
		t.Fatalf("Session.ID=%q want %q", sp.ID, req.SessionID)
	}
	assertPromptCount(t, sp, 2)
	assertPromptText(t, sp, 0, "hello world")
	assertPromptText(t, sp, 1, "second prompt")
	assertPromptTimeUTC(t, sp, 0, atFixed(-30*time.Minute))
	assertPromptTimeUTC(t, sp, 1, atFixed(-15*time.Minute))
	// Optional 1-based Index if implementer fills it
	if sp.UserPrompts[0].Index != 0 && sp.UserPrompts[0].Index != 1 {
		t.Fatalf("Index[0]=%d want 0 (unset) or 1", sp.UserPrompts[0].Index)
	}
	if sp.UserPrompts[0].Index == 1 && sp.UserPrompts[1].Index != 2 {
		t.Fatalf("Index[1]=%d want 2 when 1-based", sp.UserPrompts[1].Index)
	}
}
```
