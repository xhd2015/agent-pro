## Expected

- Three sessions returned in descending `started_at` order:
  `...000003`, `...000002`, `...000001`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	want := []string{
		"01900003-0000-7000-8000-000000000003",
		"01900003-0000-7000-8000-000000000002",
		"01900003-0000-7000-8000-000000000001",
	}
	if len(resp.Sessions) != len(want) {
		t.Fatalf("len(sessions) = %d, want %d", len(resp.Sessions), len(want))
	}
	for i, id := range want {
		if resp.Sessions[i].ID != id {
			t.Fatalf("sessions[%d].ID = %q, want %q", i, resp.Sessions[i].ID, id)
		}
	}
}
```