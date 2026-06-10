## Expected
- Validation fails with a stable event type error.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.Error != "event_type is required" { t.Fatalf("error = %q", resp.Error) }
}
```

