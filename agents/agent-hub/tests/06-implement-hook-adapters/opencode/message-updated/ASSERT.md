```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { if resp.EventType!="agent.prompt.submitted" || resp.Prompt!="hello opencode" { t.Fatalf("resp=%+v", resp) } }
```

