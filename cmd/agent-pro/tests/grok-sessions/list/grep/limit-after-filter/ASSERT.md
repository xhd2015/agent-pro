## Expected

- Exactly 2 sessions returned (limit after filter).
- Order newest match first:
  1. `01900016-0000-7000-8000-000000000004`
  2. `01900016-0000-7000-8000-000000000003`
- Non-matching newest session `...000005` is absent.
- Older matches `...000002` and `...000001` are absent due to limit.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	want := []string{
		"01900016-0000-7000-8000-000000000004",
		"01900016-0000-7000-8000-000000000003",
	}
	if len(resp.Sessions) != len(want) {
		t.Fatalf("len(sessions) = %d, want %d; sessions=%v", len(resp.Sessions), len(want), sessionIDs(resp))
	}
	for i, id := range want {
		if resp.Sessions[i].ID != id {
			t.Fatalf("sessions[%d].ID = %q, want %q", i, resp.Sessions[i].ID, id)
		}
	}
	assertNotContains(t, resp.Output, "01900016-0000-7000-8000-000000000005")
	assertNotContains(t, resp.Output, "01900016-0000-7000-8000-000000000002")
	assertNotContains(t, resp.Output, "01900016-0000-7000-8000-000000000001")
	assertContains(t, resp.Output, "GREP_LIMIT_TOKEN")
}

func sessionIDs(resp *Response) []string {
	ids := make([]string, len(resp.Sessions))
	for i, s := range resp.Sessions {
		ids[i] = s.ID
	}
	return ids
}
```
