## Expected
- Validation fails with a stable unknown event error.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if !strings.Contains(resp.Error, "unknown event_type") { t.Fatalf("error = %q", resp.Error) }
}
```

