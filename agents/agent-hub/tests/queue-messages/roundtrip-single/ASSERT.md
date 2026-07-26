## Expected
- Single message returned.
- text:"exact message", runner_session_id matches.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "message", "pop", "--runner", "fake-opencode", "--session-id", "srt1")
    if err != nil {
        t.Fatalf("pop error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    msgs, _ := obj["messages"].([]any)
    if msgs == nil || len(msgs) != 1 {
        t.Fatalf("expected 1 message, got %v", len(msgs))
    }
    m := msgs[0].(map[string]any)
    if m["text"] != "exact message" {
        t.Fatalf("expected 'exact message', got %v", m["text"])
    }
    if m["session_id"] != "srt1" {
        t.Fatalf("expected session_id srt1, got %v", m["session_id"])
    }
}
```
