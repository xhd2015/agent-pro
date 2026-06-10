## Expected
- The hook survives JSON round trip.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil || !resp.OK { t.Fatalf("json=%s resp=%+v err=%v", resp.JSON, resp, err) }
}
```

