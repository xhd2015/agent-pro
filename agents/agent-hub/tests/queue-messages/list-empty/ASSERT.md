## Expected
- {"messages":[]}.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "message", "list", "--runner", "fake-opencode", "--session-id", "sempty")
    if err != nil {
        t.Fatalf("list error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    msgs, _ := obj["messages"].([]any)
    if msgs == nil || len(msgs) != 0 {
        t.Fatalf("expected empty messages, got %v", len(msgs))
    }
}
```
