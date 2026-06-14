## Steps
1. Write a plugin that logs to a marker file when `session.created` fires.
2. Run `fake-opencode run --plugin <plugin.ts> --mock-config <config>`.
3. Verify the marker file was written with expected content.

```go
import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    markerPath := filepath.Join(req.TempDir, "handler-called.json")
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
    pluginPath := filepath.Join(req.TempDir, "plugin.ts")
    writeFile(t, pluginPath, pluginContent)

    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_plugin","llm_events":[{"type":"message","done":true}]}`)

    req.MarkerPath = markerPath
    req.Args = []string{"run", "--format", "json", "--mock-config", req.MockConfigPath, "--plugin", pluginPath, "hello"}
    return nil
}
```
