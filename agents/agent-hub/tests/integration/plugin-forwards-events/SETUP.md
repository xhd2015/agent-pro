## Steps
1. Install a test plugin via `agent-hub integration install opencode`.
2. Overwrite the installed plugin with a version that calls `agent-hub hook notify` for session events.
3. Set up Run to execute fake-opencode with the plugin.
4. In Assert, fetch events from agent-hub to verify they arrived.

```go
import (
    "fmt"
    "os/exec"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    if _, err := exec.LookPath("bun"); err != nil {
        t.Skipf("skipping plugin test: bun not installed")
    }

    opencodeHome := filepath.Join(req.TempDir, "opencode-home")
    req.Env = append(req.Env, "OPENCODE_CONFIG_DIR="+opencodeHome)

    instResp, err := runAgentHub(t, req, "integration", "opencode", "install", "--opencode-home", opencodeHome)
    if err != nil {
        return err
    }
    if instResp.ExitCode != 0 {
        return fmt.Errorf("install failed: %s", instResp.Stderr)
    }

    pluginPath := filepath.Join(opencodeHome, "plugins", "agent-hub.ts")
    pluginContent := fmt.Sprintf(`import { execSync } from "child_process";
const agentHub = %q;
function notify(eventType, payload) {
  execSync(agentHub + " hook notify --runner opencode --event " + eventType, {
    input: JSON.stringify(payload),
    stdio: ["pipe", "pipe", "pipe"],
    env: process.env,
  });
}
export const AgentHubPlugin = async () => {
  return {
    "session.created": async (event) => {
      notify("session.created", event);
    },
  };
};
`, req.AgentHub)
    writeFile(t, pluginPath, pluginContent)

    mockConfigPath := filepath.Join(req.TempDir, "mock-e2e.json")
    mockConfig := `{"version":"agent-pro.fake-runner.v1","runner":"opencode","session_id":"sess_e2e","llm_events":[{"type":"message","text":"hello"},{"type":"done"}]}`
    writeFile(t, mockConfigPath, mockConfig)

    req.Command = req.FakeOpencode
    req.Args = []string{"run", "--format", "json", "--mock-config", mockConfigPath, "--plugin", pluginPath, "hello"}

    return nil
}
```
