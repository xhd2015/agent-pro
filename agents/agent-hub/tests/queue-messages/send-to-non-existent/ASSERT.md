## Expected
- session_status:"running" (auto-creates the session).
- session show returns status:"running" with correct runner_session_id.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "newone", "--text", "hello")
    if err != nil {
        t.Fatalf("send error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    if obj["session_status"] != "running" {
        t.Fatalf("expected session_status running, got %v", obj["session_status"])
    }

    sr, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "newone")
    if err != nil {
        t.Fatalf("show error: %v", err)
    }
    assertSuccess(t, sr)
    so := parseJSON(t, sr.Stdout)
    if so["status"] != "running" {
        t.Fatalf("expected status running, got %v", so["status"])
    }
    if so["runner_session_id"] != "newone" {
        t.Fatalf("expected runner_session_id newone, got %v", so["runner_session_id"])
    }
}
```
