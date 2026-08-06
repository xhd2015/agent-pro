## Expected

- No error.
- Exactly 2 sessions (A then B).
- Each session has exactly 1 prompt whose text contains `keepme`.
- No `ignore-other` or lone `noise` texts remain.

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
	assertListIDsNewestFirst(t, resp.List, []string{idFilterGrepA, idFilterGrepB})
	for i, sp := range resp.List {
		if len(sp.UserPrompts) != 1 {
			t.Fatalf("list[%d] prompts=%d want 1: %+v", i, len(sp.UserPrompts), sp.UserPrompts)
		}
		if !strings.Contains(sp.UserPrompts[0].Text, "keepme") {
			t.Fatalf("list[%d] prompt %q missing keepme", i, sp.UserPrompts[0].Text)
		}
	}
}
```
