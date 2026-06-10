## Expected
- Validation succeeds.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil || !resp.OK { t.Fatalf("valid minimal failed: err=%v resp=%+v", err, resp) }
}
```

