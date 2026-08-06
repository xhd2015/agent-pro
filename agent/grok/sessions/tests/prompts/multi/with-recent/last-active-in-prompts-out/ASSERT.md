## Expected

- No error.
- Exactly 1 session: multiSessionID(1) (`true in-window`).
- `idLastActiveOnly` is **not** in the list (prompt ts outside window).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	for _, sp := range resp.List {
		if sp.ID == idLastActiveOnly {
			t.Fatalf("session %s must be skipped (prompts outside window)", idLastActiveOnly)
		}
	}
	assertListIDsNewestFirst(t, resp.List, []string{multiSessionID(1)})
	assertPromptCount(t, &resp.List[0], 1)
	assertPromptText(t, &resp.List[0], 0, "fresh")
}
```
