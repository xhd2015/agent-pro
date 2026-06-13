## Expected
- Response has message and session_status:"running".
- session show still returns status:"running".

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "s1", "--text", "followup")
    if err != nil {
        t.Fatalf("send error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    if obj["session_status"] != "running" {
        t.Fatalf("expected session_status running, got %v", obj["session_status"])
    }
    if _, ok := obj["message"]; !ok {
        t.Fatal("message field missing")
    }

    sr, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "s1")
    if err != nil {
        t.Fatalf("show error: %v", err)
    }
    assertSuccess(t, sr)
    so := parseJSON(t, sr.Stdout)
    if so["status"] != "running" {
        t.Fatalf("expected status running, got %v", so["status"])
    }
}
```
