## Expected
- `err` is non-nil.
- The error message contains "claude" (case-insensitive).

## Side Effects
- None (no real claude process is spawned).

## Exit Code
- Not applicable (in-process agent call, not a CLI invocation).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected non-nil error when claude binary path does not exist")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "claude") {
		t.Fatalf("expected error to mention 'claude', got: %v", err)
	}
}
```
