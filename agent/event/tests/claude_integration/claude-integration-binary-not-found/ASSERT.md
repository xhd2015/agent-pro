## Expected
- The `runClaudeHeadless` helper returns a non-nil error because the explicitly-configured `ClaudePath` does not exist.
- The error message references the missing binary path or contains "claude".

## Side Effects
- No claude process is started.

## Errors
- A non-nil error is returned.
- If `CLAUDE_SKIP_INTEGRATION=1` is set the test may skip (user opted out).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected non-nil error for nonexistent ClaudePath")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "claude") && !strings.Contains(msg, "nonexistent") {
		t.Fatalf("expected error to mention claude or nonexistent path, got: %v", err)
	}
}
```
