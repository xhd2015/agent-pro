## Expected
- Returns {"messages":[...]} with 2 items.
- Second pop returns {"messages":[]} (drained).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "message", "pop", "--runner", "fake-opencode", "--session-id", "spop")
    if err != nil {
        t.Fatalf("pop error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    msgs, _ := obj["messages"].([]any)
    if msgs == nil || len(msgs) != 2 {
        t.Fatalf("expected 2 messages, got %v", len(msgs))
    }

    r2, err := runAgentHub(t, req, "session", "message", "pop", "--runner", "fake-opencode", "--session-id", "spop")
    if err != nil {
        t.Fatalf("pop error: %v", err)
    }
    assertSuccess(t, r2)
    obj2 := parseJSON(t, r2.Stdout)
    msgs2, _ := obj2["messages"].([]any)
    if msgs2 == nil || len(msgs2) != 0 {
        t.Fatalf("expected empty messages after second pop, got %v", len(msgs2))
    }
}
```
