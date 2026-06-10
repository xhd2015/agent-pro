```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { if !resp.OK { t.Fatalf("missing %q", resp.Want) } }
```

