## Expected
- Stderr references `my_sid_custom` (from the custom env var).
- The custom `SessionEnvVar` was used to resolve the session ID.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if !strings.Contains(resp.Stderr, "my_sid_custom") {
        t.Fatalf("expected 'my_sid_custom' in stderr (custom env var used), got:\n%s", resp.Stderr)
    }
    if !strings.Contains(resp.Stderr, "session not found") {
        t.Fatalf("expected 'session not found' in stderr, got:\n%s", resp.Stderr)
    }
}
```
