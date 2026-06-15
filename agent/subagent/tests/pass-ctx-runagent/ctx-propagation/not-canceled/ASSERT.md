## Expected
- `runAgent` returns with no error when called with a normal `context.Background()`.
- The output is non-empty, indicating the agent executed successfully.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("expected no error from normal context, got: %v", err)
    }
    if strings.TrimSpace(resp.Output) == "" {
        t.Fatal("expected non-empty output from agent, got empty")
    }
}
```
