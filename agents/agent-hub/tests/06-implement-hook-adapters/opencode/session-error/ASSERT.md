```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.EventType!="agent.session.failed" { t.Fatalf("resp=%+v", resp) } }
```

