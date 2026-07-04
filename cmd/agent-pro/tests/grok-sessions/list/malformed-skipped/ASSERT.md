## Expected

- `List` returns exactly 2 valid sessions.
- Newest session is `01900005-0000-7000-8000-000000000002`.
- Malformed files do not appear in results and do not cause an error.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 2 {
		t.Fatalf("len(sessions) = %d, want 2", len(resp.Sessions))
	}
	want := []string{
		"01900005-0000-7000-8000-000000000002",
		"01900005-0000-7000-8000-000000000001",
	}
	for i, id := range want {
		if resp.Sessions[i].ID != id {
			t.Fatalf("sessions[%d].ID = %q, want %q", i, resp.Sessions[i].ID, id)
		}
	}
}
```