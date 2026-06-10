## Expected
- Three events are returned.

```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.Count != 3 || resp.Cursor.Offset != 3 { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

