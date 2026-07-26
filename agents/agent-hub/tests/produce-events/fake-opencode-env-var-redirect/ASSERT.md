## Expected
- Event stored under runner "fake-opencode" (not "opencode").
- Session directory exists under sessions/fake-opencode/.

```go
import (
    "testing"
    "path/filepath"
    "os"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
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
    if ev["runner"] != "fake-opencode" {
        t.Fatalf("expected runner fake-opencode after redirect, got %v", ev["runner"])
    }

    // check sessions/fake-opencode/ directory exists
    sessionDir := filepath.Join(req.Home, "sessions", "fake-opencode")
    if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
        t.Fatalf("sessions/fake-opencode directory not found")
    }
}
```
