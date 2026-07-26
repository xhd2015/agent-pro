## Expected
- {"messages":[]}.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "message", "pop", "--runner", "fake-opencode", "--session-id", "spop_empty")
    if err != nil {
        t.Fatalf("pop error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    msgs, _ := obj["messages"].([]any)
    if msgs == nil || len(msgs) != 0 {
        t.Fatalf("expected empty messages, got %v", len(msgs))
    }
}
```
