## Expected
- Two events are returned and cursor advances to offset two.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) { if len(resp.Events) != 2 || resp.Cursor.Offset != 2 { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

