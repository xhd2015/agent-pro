```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.EventType!="agent.tool.finished" { t.Fatalf("resp=%+v", resp) } }
```

