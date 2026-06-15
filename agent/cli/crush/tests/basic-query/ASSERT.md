## Expected
- Response.Answer contains "paris" (case-insensitive).
- Response.SessionID is non-empty (a session was created by the crush server).

## Side Effects
- None.

## Exit Code
- Not applicable (in-process agent call, not a CLI invocation).

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }
    if resp.SessionID == "" {
        t.Fatal("expected non-empty SessionID after Ask()")
    }
    lower := strings.ToLower(resp.Answer)
    if !strings.Contains(lower, "paris") {
        t.Fatalf("expected answer to contain 'paris', got: %s", resp.Answer)
    }
}
```
