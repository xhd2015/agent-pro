## Expected
- The storage partition follows received time.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) { if !resp.OK || resp.Partitions[0] != "2026-06-11" { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

