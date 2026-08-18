## Expected

- `NormalizeIdle(true, -1s)` returns a non-empty error (negative duration).
- `BuildFollowUpCommand` also returns an API error when normalize sits on the
  emit path.
- No requirement that a follow-up line is produced.

## Side Effects

- None (pure).

## Errors

- Normalize error required. Emit error required (normalize on emit path).

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.NormalizeErr == "" {
		t.Fatal("NormalizeIdle(true, -1s) must return an error")
	}
	if resp.ErrString == "" {
		t.Fatal("BuildFollowUpCommand must return an API error for negative IdleTimeout when ExitOnIdle is set")
	}
	if resp.Enabled {
		t.Fatal("NormalizeIdle error path must not report enabled=true")
	}
}
```
