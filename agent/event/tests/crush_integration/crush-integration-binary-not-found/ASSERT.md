## Expected
- The Run function returns a non-nil error because the explicitly-configured
  `CrushPath` does not exist.
- The error message references the missing binary path or contains "crush".

## Side Effects
- No server process is started.

## Errors
- A non-nil error is returned.
- If `CRUSH_SKIP_INTEGRATION=1` is set the test may skip (user opted out).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected non-nil error for nonexistent CrushPath")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "crush") && !strings.Contains(msg, "nonexistent") {
		t.Fatalf("expected error to mention crush or nonexistent path, got: %v", err)
	}
}
```
