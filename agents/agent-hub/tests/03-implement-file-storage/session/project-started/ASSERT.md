## Expected
- Active session file exists.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) { if !resp.OK { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

