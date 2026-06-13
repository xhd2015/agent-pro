## Steps
1. Install a test plugin via `agent-hub integration install opencode`.
2. Overwrite the installed plugin with a version that calls `agent-hub hook notify` for session events.
3. Set up Run to execute fake-opencode with the plugin.
4. In Assert, fetch events from agent-hub to verify they arrived.

```go
import (
    "os/exec"
    "fmt"
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    if _, err := exec.LookPath("bun"); err != nil {
        t.Skipf("skipping plugin test: bun not installed")
    }

    resp, err := runAgentHub(t, req, "integration", "opencode", "install")
    if err != nil {
        return err
    }
    if resp.ExitCode != 0 {
        return fmt.Errorf("install failed: %s", resp.Stderr)
    }

    pluginPath := filepath.Join(req.TempDir, ".opencode", "plugins", "agent-hub.ts")
    pluginContent := `const { execSync } = require("child_process");
function notify(eventType, payload) {
  try {
    execSync("agent-hub hook notify --runner opencode --event " + eventType, { input: JSON.stringify(payload), stdio: ["pipe","pipe","pipe"] });
  } catch(e) {}
}
export const AgentHubPlugin = async () => {
  return {
    "session.created": async (event) => {
      notify("session.created", event);
    },
  };
};
`
    writeFile(t, pluginPath, pluginContent)

    mockConfigPath := filepath.Join(req.TempDir, "mock-e2e.json")
    mockConfig := `{"version":"agent-pro.fake-runner.v1","runner":"opencode","session_id":"sess_e2e","stdout_events":[{"type":"message","done":true}]}`
    writeFile(t, mockConfigPath, mockConfig)

    req.Command = req.FakeOpencode
    req.Args = []string{"run", "--format", "json", "--mock-config", mockConfigPath, "--plugin", pluginPath, "hello"}

    return nil
}
```
