## Expected
- Event stored with runner "opencode".
- Session stored under sessions/opencode/.

```go
import (
    "testing"
    "path/filepath"
    "os"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "test-"+t.Name(), "--limit", "10")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    fr := parseJSON(t, r.Stdout)
    events := fr["events"].([]any)
    if events == nil || len(events) < 1 {
        t.Fatal("expected at least 1 event")
    }
    ev := events[0].(map[string]any)["event"].(map[string]any)
    if ev["runner"] != "opencode" {
        t.Fatalf("expected runner opencode, got %v", ev["runner"])
    }

    sessionDir := filepath.Join(req.Home, "sessions", "opencode")
    if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
        t.Fatalf("sessions/opencode directory not found")
    }
}
```
