## Expected

- `List` returns exactly 20 sessions without error.
- Returned sessions are the 20 newest by `time.updated` (minutes 5–24 of the fixture range).
- Oldest fixture `ses_limit_01` through `ses_limit_05` are excluded.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 20 {
		t.Fatalf("len(sessions) = %d, want 20", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "ses_limit_25" {
		t.Fatalf("sessions[0].ID = %q, want ses_limit_25", resp.Sessions[0].ID)
	}
	if resp.Sessions[19].ID != "ses_limit_06" {
		t.Fatalf("sessions[19].ID = %q, want ses_limit_06", resp.Sessions[19].ID)
	}
	for _, id := range []string{"ses_limit_01", "ses_limit_02", "ses_limit_03", "ses_limit_04", "ses_limit_05"} {
		for _, s := range resp.Sessions {
			if s.ID == id {
				t.Fatalf("unexpected old session %q in result", id)
			}
		}
	}
}
```