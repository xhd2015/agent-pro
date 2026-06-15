## Preconditions
- `bun` is installed on the system (test fails if not found).
- `OPENCODE_CONFIG_DIR` and `AGENT_HUB_HOME` point to isolated temp directories.
- `AGENT_HUB_OPENCODE_RUNNER=fake-opencode` is set so hook events are stored under the fake-opencode runner.

## Steps
1. Install the agent-hub plugin via `--opencode-home` into fake-opencode's config home.
2. Overwrite the installed plugin with a test version that forwards `session.created`, `session.idle`, and `session.error` via `agent-hub hook notify`.
3. Write mock config run1: [step_start, sleep(3s), message "working on it", done].
4. Write mock config run2 (resume): [message "resumed work"].
5. The Run function starts run1 in the background, performs mid-flight checks (status=running, first events fetchable), waits for completion, and returns the response.

```go
import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"
)

func Setup(t *testing.T, req *Request) error {
    if _, err := exec.LookPath("bun"); err != nil {
        t.Fatalf("full-workflow test requires bun: %v", err)
    }

    req.Env = append(req.Env,
        "AGENT_HUB_OPENCODE_RUNNER=fake-opencode",
    )

    opencodeHome := filepath.Join(req.TempDir, "opencode-home")
    req.Env = append(req.Env, "OPENCODE_CONFIG_DIR="+opencodeHome)

    instResp, err := runAgentHub(t, req, "integration", "opencode", "install", "--opencode-home", opencodeHome)
    if err != nil {
        return fmt.Errorf("install plugin: %w", err)
    }
    if instResp.ExitCode != 0 {
        return fmt.Errorf("install plugin failed (exit %d): %s", instResp.ExitCode, instResp.Stderr)
    }

    pluginPath := filepath.Join(opencodeHome, "plugins", "agent-hub.ts")
    testPlugin := `const { execSync } = require("child_process");
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
    "session.idle": async (event) => {
      notify("session.idle", event);
    },
    "session.error": async (event) => {
      notify("session.error", event);
    },
  };
};
`
    writeFile(t, pluginPath, testPlugin)

    writeFile(t, filepath.Join(req.TempDir, "run1-mock.json"), `{"version":"agent-pro.fake-runner.v1","runner":"opencode","session_id":"sess_full","llm_events":[{"type":"step_start"},{"type":"sleep","delay_ms":3000},{"type":"message","text":"working on it"},{"type":"done"}]}`)

    writeFile(t, filepath.Join(req.TempDir, "run2-mock.json"), `{"version":"agent-pro.fake-runner.v1","runner":"opencode","session_id":"sess_full","llm_events":[{"type":"message","text":"resumed work"}]}`)

    req.Operation = "full_workflow"
    req.Command = req.FakeOpencode
    req.Args = []string{"run", "--format", "json", "--mock-config", filepath.Join(req.TempDir, "run1-mock.json"), "--plugin", pluginPath, "do the task"}

    return nil
}
```
