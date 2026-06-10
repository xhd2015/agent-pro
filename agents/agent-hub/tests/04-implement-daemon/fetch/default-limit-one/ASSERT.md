## Expected
- One event is returned.

```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.Count != 1 || resp.Cursor.Offset != 1 { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

