## Steps
1. Write a plugin whose handler throws an Error.
2. Run fake-opencode.
3. Verify it exits 0 (non-fatal) and stderr contains error info.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    pluginContent := `
export const AgentHubPlugin = async () => {
  return {
    "session.created": async (event) => {
      throw new Error("plugin handler failed");
    },
  };
};
`
    pluginPath := filepath.Join(req.TempDir, "plugin.ts")
    writeFile(t, pluginPath, pluginContent)

    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_err","llm_events":[{"type":"message","done":true}],"plugins":["`+pluginPath+`"]}`)

    req.Args = []string{"run", "--format", "json", "--mock-config", req.MockConfigPath, "--plugin", pluginPath, "hello"}
    return nil
}
```
