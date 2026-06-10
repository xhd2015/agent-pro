## Expected
- Payload JSON is preserved.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if !resp.OK || resp.JSON != `{"nested":{"value":1}}` { t.Fatalf("resp=%+v err=%v", resp, err) }
}
```

