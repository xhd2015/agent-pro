## Expected

- `List` returns exactly 20 sessions (not 25).
- The oldest returned session corresponds to minute 5 (`01900001-0000-7000-8000-000000000006`).
- The newest returned session corresponds to minute 24.

## Errors

- None.

## Exit Code

- Success (no error from List).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 20 {
		t.Fatalf("len(sessions) = %d, want 20", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "01900001-0000-7000-8000-000000000025" {
		t.Fatalf("newest session id = %q, want 01900001-0000-7000-8000-000000000025", resp.Sessions[0].ID)
	}
	if resp.Sessions[len(resp.Sessions)-1].ID != "01900001-0000-7000-8000-000000000006" {
		t.Fatalf("oldest kept id = %q, want 01900001-0000-7000-8000-000000000006", resp.Sessions[len(resp.Sessions)-1].ID)
	}
}
```