## Expected
- Validation fails with partition required.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.Error != "partition is required" { t.Fatalf("error = %q", resp.Error) }
}
```

