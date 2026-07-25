## Expected

- `err != nil`.
- Error message contains `pid not found`.
- Prefer no successful Result; if Result is non-nil it must not claim a hard hit
  (`Kind` must not be `grok` or `codex` with Confidence=hard). Primary contract
  is the error.

## Side Effects

- None.

## Errors

- `pid not found` (required substring).

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err == nil {
		t.Fatal("expected error for unknown pid, got nil")
	}
	if !strings.Contains(err.Error(), "pid not found") {
		t.Fatalf("error: got %q, want substring %q", err.Error(), "pid not found")
	}
	// Optional: if a Result was still returned, it must not be a hard session hit.
	if resp != nil && resp.Result != nil {
		r := resp.Result
		if r.Confidence == "hard" || r.Kind == "grok" || r.Kind == "codex" {
			t.Fatalf("on unknown pid, must not return hard hit: %+v", r)
		}
	}
}
```
