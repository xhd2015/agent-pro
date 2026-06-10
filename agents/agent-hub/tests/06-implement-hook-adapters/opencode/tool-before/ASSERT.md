```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.EventType!="agent.tool.started" { t.Fatalf("resp=%+v", resp) } }
```

