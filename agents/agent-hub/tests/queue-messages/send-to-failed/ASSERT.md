## Expected
- session_status:"running".
- session show returns status:"running".

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "s3", "--text", "retry")
    if err != nil {
        t.Fatalf("send error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    if obj["session_status"] != "running" {
        t.Fatalf("expected session_status running, got %v", obj["session_status"])
    }

    sr, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "s3")
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
