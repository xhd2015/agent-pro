## Expected
- First show returns status:"running".
- After finished event, show returns status:"completed".
- Both have same runner_session_id.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "s_trans")
    if err != nil {
        t.Fatalf("session show error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    if obj["status"] != "running" {
        t.Fatalf("expected status running, got %v", obj["status"])
    }

    // transition to finished
    notifyEvent(t, req, "agent.session.finished", "fake-opencode", "s_trans")

    r2, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "s_trans")
    if err != nil {
        t.Fatalf("session show error: %v", err)
    }
    assertSuccess(t, r2)
    obj2 := parseJSON(t, r2.Stdout)
    if obj2["status"] != "completed" {
        t.Fatalf("expected status completed, got %v", obj2["status"])
    }
    if obj2["runner_session_id"] != "s_trans" {
        t.Fatalf("expected runner_session_id s_trans, got %v", obj2["runner_session_id"])
    }
}
```
