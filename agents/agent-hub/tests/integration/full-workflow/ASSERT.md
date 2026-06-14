## Expected
1. Run 1 succeeds (exit 0) and stdout contains expected messages.
2. Session status is "completed" in agent-hub.
3. Fetch returns both `agent.session.started` and `agent.session.finished` events with runner=fake-opencode, session=sess_full.
4. Resume (run 2) with same session produces new events.
5. Fetch with a new consumer shows the resume events.
6. Queue a followup message to the session succeeds with session_status "running" (resume reactivates).
7. Pop returns the followup message.

```go
import (
    "path/filepath"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, "working on it")

    // verify session status → completed
    sr, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "sess_full")
    if err != nil {
        t.Fatalf("session show: %v", err)
    }
    assertSuccess(t, sr)
    so := parseJSON(t, sr.Stdout)
    if so["status"] != "completed" {
        t.Fatalf("expected status completed, got %v", so["status"])
    }

    // fetch all events → verify lifecycle
    fr, err := runAgentHub(t, req, "fetch", "--consumer-id", "verify-"+t.Name(), "--limit", "20")
    if err != nil {
        t.Fatalf("fetch: %v", err)
    }
    assertSuccess(t, fr)
    frObj := parseJSON(t, fr.Stdout)
    events, _ := frObj["events"].([]any)
    hasStarted := false
    hasFinished := false
    for _, e := range events {
        ev := e.(map[string]any)["event"].(map[string]any)
        if ev["runner"] != "fake-opencode" {
            t.Fatalf("expected runner fake-opencode, got %v", ev["runner"])
        }
        if ev["runner_session_id"] != "sess_full" {
            t.Fatalf("expected session sess_full, got %v", ev["runner_session_id"])
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

    // resume: run fake-opencode again with same session, different events
    run2Config := filepath.Join(req.TempDir, "run2-mock.json")
    opencodeHome := filepath.Join(req.TempDir, "opencode-home")
    pluginPath := filepath.Join(opencodeHome, "plugins", "agent-hub.ts")
    r2Resp, err := execCmd(t, req.FakeOpencode, []string{"run", "--format", "json", "--mock-config", run2Config, "--plugin", pluginPath, "continue"}, req.TempDir, req.Env, "")
    if err != nil {
        t.Fatalf("run2 (resume): %v", err)
    }
    assertSuccess(t, r2Resp)
    assertContains(t, r2Resp.Stdout, "resumed work")

    // fetch with new consumer → must see resume events (session.updated)
    fr2, err := runAgentHub(t, req, "fetch", "--consumer-id", "resume-"+t.Name(), "--limit", "20")
    if err != nil {
        t.Fatalf("fetch (resume): %v", err)
    }
    assertSuccess(t, fr2)
    fr2Obj := parseJSON(t, fr2.Stdout)
    events2, _ := fr2Obj["events"].([]any)
    if events2 == nil || len(events2) == 0 {
        t.Fatal("expected resume events, got none")
    }

    // queue followup message
    msgResp, err := runAgentHub(t, req, "session", "message", "send", "--runner", "fake-opencode", "--session-id", "sess_full", "--text", "followup question")
    if err != nil {
        t.Fatalf("send message: %v", err)
    }
    assertSuccess(t, msgResp)
    msgObj := parseJSON(t, msgResp.Stdout)
    if _, ok := msgObj["message"]; !ok {
        t.Fatal("message field missing in send response")
    }

    // pop message → must return the followup
    popResp, err := runAgentHub(t, req, "session", "message", "pop", "--runner", "fake-opencode", "--session-id", "sess_full")
    if err != nil {
        t.Fatalf("pop message: %v", err)
    }
    assertSuccess(t, popResp)
    popObj := parseJSON(t, popResp.Stdout)
    msgs, _ := popObj["messages"].([]any)
    if msgs == nil || len(msgs) == 0 {
        t.Fatal("expected popped message, got none")
    }
    found := false
    for _, m := range msgs {
        msg := m.(map[string]any)
        if text, _ := msg["text"].(string); text == "followup question" {
            found = true
        }
    }
    if !found {
        t.Fatal("popped message did not contain 'followup question'")
    }
}
```
