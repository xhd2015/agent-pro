## Expected
- The new day starts at offset zero.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.Offsets[0] != 0 || resp.Partitions[0] != "2026-06-11" { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

