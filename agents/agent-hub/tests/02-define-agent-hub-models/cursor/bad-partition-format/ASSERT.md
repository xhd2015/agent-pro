## Expected
- Validation fails with partition format error.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.Error != "partition must use YYYY-MM-DD" { t.Fatalf("error = %q", resp.Error) }
}
```

