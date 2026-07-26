## Expected

- Three sessions returned in descending `time.updated` order:
  `ses_sort_03`, `ses_sort_02`, `ses_sort_01`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	want := []string{"ses_sort_03", "ses_sort_02", "ses_sort_01"}
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