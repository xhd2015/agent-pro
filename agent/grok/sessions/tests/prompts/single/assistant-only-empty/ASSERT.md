## Expected

- No error (session found).
- `UserPrompts` length 0.
- Session.ID still populated.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.Single == nil {
		t.Fatal("Single is nil")
	}
	if resp.Single.ID != req.SessionID {
		t.Fatalf("ID=%q want %q", resp.Single.ID, req.SessionID)
	}
	assertPromptCount(t, resp.Single, 0)
}
```
