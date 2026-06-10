## Expected
- Validation fails with a stable runner error.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.Error != "runner is required" { t.Fatalf("error = %q", resp.Error) }
}
```

