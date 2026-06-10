## Expected
- Missing days are skipped.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) { if len(resp.Events) != 1 || resp.Events[0].Partition != "2026-06-13" { t.Fatalf("resp=%+v err=%v", resp, err) } }
```

