## Expected
- fake-opencode exits 0.
- fetch returns 2 events: agent.session.started + agent.session.finished with runner:fake-opencode, runner_session_id:sess_life.
- session show returns status:completed.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    // fetch events
    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "test-"+t.Name(), "--limit", "10")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    fr := parseJSON(t, r.Stdout)
    events := fr["events"].([]any)
    if events == nil || len(events) < 2 {
        t.Fatalf("expected at least 2 events, got %v", len(events))
    }

    hasStarted := false
    hasFinished := false
    for _, e := range events {
        ev := e.(map[string]any)["event"].(map[string]any)
        if ev["runner"] != "fake-opencode" {
            t.Fatalf("expected runner fake-opencode, got %v", ev["runner"])
        }
        if ev["runner_session_id"] != "sess_life" {
            t.Fatalf("expected session sess_life, got %v", ev["runner_session_id"])
        }
        switch ev["event_type"] {
        case "agent.session.started":
            hasStarted = true
        case "agent.session.finished":
            hasFinished = true
        }
    }
    if !hasStarted {
        t.Fatal("missing agent.session.started event")
    }
    if !hasFinished {
        t.Fatal("missing agent.session.finished event")
    }

    // session show
    sr, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "sess_life")
    if err != nil {
        t.Fatalf("session show error: %v", err)
    }
    assertSuccess(t, sr)
    so := parseJSON(t, sr.Stdout)
    if so["status"] != "completed" {
        t.Fatalf("expected status completed, got %v", so["status"])
    }
}
```
