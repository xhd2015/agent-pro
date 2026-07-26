## Preconditions
- `OPENCODE_CONFIG_DIR` points to a temp directory containing `plugins/agent-hub.ts`.
- No `--plugin` flag or mock config plugins are provided.

## Steps
1. Create a plugin file at `{OPENCODE_CONFIG_DIR}/plugins/agent-hub.ts` that writes a marker on `session.created`.
2. Run fake-opencode with a mock config (no explicit plugins).
3. Verify the marker file was written (plugin was auto-discovered and fired).

```go
import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
	"os/exec"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    if _, err := exec.LookPath("bun"); err != nil {
        t.Skipf("skipping plugin test: bun not installed")
    }

    markerPath := filepath.Join(req.TempDir, "auto-discovered.json")
    opencodeConfigDir := filepath.Join(req.TempDir, "opencode-config")
    pluginsDir := filepath.Join(opencodeConfigDir, "plugins")
    if err := os.MkdirAll(pluginsDir, 0755); err != nil {
        return err
    }
    pluginContent := `
import { writeFileSync } from "fs";
export const AgentHubPlugin = async () => {
  return {
    "session.created": async (event) => {
      writeFileSync("` + markerPath + `", JSON.stringify(event));
    },
  };
};
`
    writeFile(t, filepath.Join(pluginsDir, "agent-hub.ts"), pluginContent)

    req.Env = append(req.Env, "OPENCODE_CONFIG_DIR="+opencodeConfigDir)
    req.MarkerPath = markerPath

    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_autodiscover","llm_events":[{"type":"message","done":true}]}`)
    return nil
}
```
