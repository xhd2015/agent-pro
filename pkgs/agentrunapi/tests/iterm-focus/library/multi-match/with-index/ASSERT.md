## Expected

- `FocusSession` returns nil error.
- Chosen candidate is the second match (`WindowID` `w-b` / TTY `/dev/ttys011`).
- `FocusITerm` called once with that ref.

## Errors

- None.

## Exit Code

- N/A (library)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	if resp.Chosen.Ref.WindowID != "w-b" {
		t.Fatalf("chosen WindowID = %q, want w-b (index 1); chosen=%+v candidates=%+v",
			resp.Chosen.Ref.WindowID, resp.Chosen, resp.Candidates)
	}
	if len(resp.FocusCalls) != 1 {
		t.Fatalf("FocusITerm calls = %d, want 1", len(resp.FocusCalls))
	}
	if resp.FocusCalls[0].WindowID != "w-b" {
		t.Fatalf("FocusITerm ref = %+v, want w-b", resp.FocusCalls[0])
	}
}
```
