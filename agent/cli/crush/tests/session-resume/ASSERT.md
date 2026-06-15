## Expected
- Response.Answer references the prior conversation, containing "french" or "capital".
- Response.SessionID is non-empty (a session was created and reused).

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
    if !strings.Contains(lower, "french") && !strings.Contains(lower, "capital") {
        t.Fatalf("expected response to reference prior question, got: %s", resp.Answer)
    }
}
```
