## Steps
1. Write a plugin that only handles `session.created`.
2. Send an event like `tool.execute.before` that has no handler.
3. Verify fake-opencode exits 0 without errors.

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
      // only handle session.created
    },
  };
};
`
    pluginPath := filepath.Join(req.TempDir, "plugin.ts")
    writeFile(t, pluginPath, pluginContent)

    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_unknown","stdout_events":[{"type":"message","done":true}],"plugins":["`+pluginPath+`"]}`)

    req.Args = []string{"run", "--format", "json", "--mock-config", req.MockConfigPath, "--plugin", pluginPath, "hello"}
    return nil
}
```
