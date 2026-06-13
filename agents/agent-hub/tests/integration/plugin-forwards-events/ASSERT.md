## Expected
- fake-opencode exits 0.
- An event with `runner: "opencode"` and `runner_session_id: "sess_e2e"` is fetchable from agent-hub.

```go
import (
    "encoding/json"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)

    fr, err := runAgentHub(t, req, "fetch", "--consumer-id", "e2e-"+t.Name(), "--limit", "5")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertExitCode(t, fr, 0)
    frObj := parseJSON(t, fr.Stdout)
    events, _ := frObj["events"].([]any)
    if events == nil || len(events) == 0 {
        t.Fatal("no events in agent-hub; plugin did not forward")
    }

    for _, evt := range events {
        e := evt.(map[string]any)
        ev, _ := e["event"].(map[string]any)
        if ev == nil {
            continue
        }
        if ev["runner"] == "opencode" && ev["runner_session_id"] == "sess_e2e" {
            return
        }
    }
    t.Fatal("expected event with runner=opencode runner_session_id=sess_e2e not found")
    _ = json.Unmarshal
}
```
