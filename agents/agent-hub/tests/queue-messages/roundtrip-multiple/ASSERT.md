## Expected
- 3 messages returned in insertion order (A, B, C).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "message", "pop", "--runner", "fake-opencode", "--session-id", "srt3")
    if err != nil {
        t.Fatalf("pop error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    msgs, _ := obj["messages"].([]any)
    if msgs == nil || len(msgs) != 3 {
        t.Fatalf("expected 3 messages, got %v", len(msgs))
    }
    expected := []string{"A", "B", "C"}
    for i, name := range expected {
        m := msgs[i].(map[string]any)
        if m["text"] != name {
            t.Fatalf("message %d: expected %q, got %v", i, name, m["text"])
        }
    }
}
```
