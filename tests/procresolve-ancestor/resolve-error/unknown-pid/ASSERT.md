## Expected

- `err != nil` from `ResolveFromAncestors`.
- Error message contains `pid not found`.
- `FindAncestorGrok` ok=false.
- Must not return a hard grok/codex hit.

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
	assertNoAncestor(t, resp)
	if resp != nil && resp.Result != nil {
		r := resp.Result
		if r.Confidence == "hard" || r.Kind == "grok" || r.Kind == "codex" {
			t.Fatalf("on unknown pid, must not return hard hit: %+v", r)
		}
	}
}
```
