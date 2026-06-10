## Expected
- Loaded cursor equals saved cursor.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) { if !resp.OK || resp.Cursor.Offset != 7 { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

