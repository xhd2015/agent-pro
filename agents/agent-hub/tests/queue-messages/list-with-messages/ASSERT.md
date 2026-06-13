## Expected
- Returns {"messages":[...]} with 2 items.
- text fields are "msg1" and "msg2".
- Second list returns same 2 items (peek, not drain).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "message", "list", "--runner", "fake-opencode", "--session-id", "slist")
    if err != nil {
        t.Fatalf("list error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    msgs, _ := obj["messages"].([]any)
    if msgs == nil || len(msgs) != 2 {
        t.Fatalf("expected 2 messages, got %v", len(msgs))
    }
    m1 := msgs[0].(map[string]any)
    m2 := msgs[1].(map[string]any)
    if m1["text"] != "msg1" {
        t.Fatalf("expected msg1, got %v", m1["text"])
    }
    if m2["text"] != "msg2" {
        t.Fatalf("expected msg2, got %v", m2["text"])
    }

    // list again - should be idempotent
    r2, err := runAgentHub(t, req, "session", "message", "list", "--runner", "fake-opencode", "--session-id", "slist")
    if err != nil {
        t.Fatalf("list error: %v", err)
    }
    assertSuccess(t, r2)
    obj2 := parseJSON(t, r2.Stdout)
    msgs2, _ := obj2["messages"].([]any)
    if msgs2 == nil || len(msgs2) != 2 {
        t.Fatalf("expected 2 messages on second list, got %v", len(msgs2))
    }
}
```
