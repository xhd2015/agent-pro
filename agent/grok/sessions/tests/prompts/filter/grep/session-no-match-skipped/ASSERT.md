## Expected

- No error.
- Exactly 2 sessions: A then C (B skipped as zero survivors).
- B id must not appear.
- Limit was not consumed by B (C still returned).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertListIDsNewestFirst(t, resp.List, []string{idFilterGrepA, idFilterGrepC})
	for _, sp := range resp.List {
		if sp.ID == idFilterGrepB {
			t.Fatal("session with no grep match must be skipped")
		}
	}
}
```
