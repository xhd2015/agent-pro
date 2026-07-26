## Expected
- ExitCode 0.
- JSON has status:"running", runner:"fake-opencode", runner_session_id:"s_run".
- last_event has partition and offset.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "s_run")
    if err != nil {
        t.Fatalf("session show error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    if obj["status"] != "running" {
        t.Fatalf("expected status running, got %v", obj["status"])
    }
    if obj["runner"] != "fake-opencode" {
        t.Fatalf("expected runner fake-opencode, got %v", obj["runner"])
    }
    if obj["runner_session_id"] != "s_run" {
        t.Fatalf("expected runner_session_id s_run, got %v", obj["runner_session_id"])
    }
    if le, ok := obj["last_event"].(map[string]any); ok {
        if le["partition"] == nil || le["offset"] == nil {
            t.Fatal("last_event missing partition or offset")
        }
    }
}
```
