## Expected
- Event remains fetchable.

```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.Count != 1 { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

