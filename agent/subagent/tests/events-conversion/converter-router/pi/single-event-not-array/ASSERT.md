## Expected
- **After fix: ConvertRawLine accepts single pi events (not wrapped in array) without error.**
- The conversion should succeed and produce correct AgentEvent output.
- The output should contain the message event with the text "Hello".

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // After fix: single pi object is accepted and converted
    if !strings.Contains(resp.Stdout, `"type":"message"`) {
        t.Fatalf("expected message event, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, "Hello") {
        t.Fatalf("expected text 'Hello', got:\n%s", resp.Stdout)
    }
}
```
