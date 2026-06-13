## Expected
- Returns JSON with 2 entries.
- Each has runner, runner_session_id, status.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "sessions")
    if err != nil {
        t.Fatalf("sessions error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    sessions, _ := obj["sessions"].([]any)
    if sessions == nil || len(sessions) < 2 {
        t.Fatalf("expected at least 2 sessions, got %v", len(sessions))
    }
    for _, s := range sessions {
        sm := s.(map[string]any)
        if sm["runner"] == nil {
            t.Fatal("session missing runner")
        }
        if sm["runner_session_id"] == nil {
            t.Fatal("session missing runner_session_id")
        }
        if sm["status"] == nil {
            t.Fatal("session missing status")
        }
    }
}
```
