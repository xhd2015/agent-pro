## Expected
- Events from both partitions are returned and cursor advances into the second partition.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) { if len(resp.Events) != 2 || resp.Events[1].Partition != "2026-06-11" || resp.Cursor.Partition != "2026-06-11" || resp.Cursor.Offset != 1 { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

