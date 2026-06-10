## Expected
- Validation fails with offset constraint.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.Error != "offset must be >= 0" { t.Fatalf("error = %q", resp.Error) }
}
```

