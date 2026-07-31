## Expected

- `FocusSession` returns nil error.
- Chosen candidate still identifies the single match (`win-1` / ttys148).
- `FocusITerm` is **never** called.

## Side Effects

- No focus side effect (mock call log empty).

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
	if resp.Chosen.Ref.WindowID != "win-1" {
		t.Fatalf("dry-run must still return chosen candidate, got %+v", resp.Chosen)
	}
	assertNoFocus(t, resp)
}
```
