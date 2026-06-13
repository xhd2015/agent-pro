## Preconditions
- Events on multiple date partitions.
- NOTE: This test requires events in at least 2 date partitions. Since we produce events on the same day, this test verifies the structure but may only find 1 partition in simple setups.

## Steps
1. Produce events via notify.
2. Also run fake-opencode with hooks to produce events.
3. Fetch with large limit to span partitions.

```go
import (
    "testing"
    "path/filepath"
)

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_cross")

    // produce more events via fake-opencode
    config := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_cross","hooks":[{"at":"before_stdout","event":"session.created"}]}`
    cfgPath := filepath.Join(req.TempDir, "mock-cross.json")
    writeFile(t, cfgPath, config)
    _, err := runFakeOpencode(t, req, cfgPath)
    if err != nil {
        t.Fatalf("fake-opencode error: %v", err)
    }
    return nil
}
```
