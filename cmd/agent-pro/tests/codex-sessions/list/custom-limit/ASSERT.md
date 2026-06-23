## Expected

- `List` returns exactly 3 sessions.
- Newest session is `01900002-0000-7000-8000-000000000005`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 3 {
		t.Fatalf("len(sessions) = %d, want 3", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "01900002-0000-7000-8000-000000000005" {
		t.Fatalf("newest id = %q, want 01900002-0000-7000-8000-000000000005", resp.Sessions[0].ID)
	}
}
```