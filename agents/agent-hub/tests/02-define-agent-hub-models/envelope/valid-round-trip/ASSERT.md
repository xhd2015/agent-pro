## Expected
- Envelope validates after JSON round trip.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil || !resp.OK { t.Fatalf("resp=%+v err=%v", resp, err) }
}
```

