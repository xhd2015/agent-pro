## Expected

- `List` returns one session with `NumDisplayEvents` (or equivalent MSGS field) equal to 5.
- `FormatListTable` output contains `MSGS` header and shows `5` for the session row.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(resp.Sessions))
	}
	if resp.Sessions[0].NumDisplayEvents != 5 {
		t.Fatalf("NumDisplayEvents = %d, want 5", resp.Sessions[0].NumDisplayEvents)
	}
	assertContains(t, resp.Output, "MSGS")
	assertContains(t, resp.Output, "01900005-aaaa-7aaa-aaaa-aaaaaaaaaaaa")
	assertContains(t, resp.Output, "5")
}
```