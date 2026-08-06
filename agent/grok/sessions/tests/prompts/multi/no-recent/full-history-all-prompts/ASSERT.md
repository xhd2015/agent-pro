## Expected

- No error.
- Exactly 1 session: idFullHistory.
- That session has **2** user prompts: `old prompt` and `recent prompt`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertListLen(t, resp.List, 1)
	if resp.List[0].ID != idFullHistory {
		t.Fatalf("ID=%q want %q", resp.List[0].ID, idFullHistory)
	}
	sp := &resp.List[0]
	assertPromptCount(t, sp, 2)
	assertPromptText(t, sp, 0, "old prompt")
	assertPromptText(t, sp, 1, "recent prompt")
}
```
