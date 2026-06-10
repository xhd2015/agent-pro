```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.EventType!="agent.session.started" { t.Fatalf("resp=%+v", resp) } }
```

