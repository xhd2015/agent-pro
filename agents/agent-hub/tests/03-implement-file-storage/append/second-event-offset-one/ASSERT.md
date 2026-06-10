## Expected
- Offset is one.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.Offsets[0] != 1 { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

