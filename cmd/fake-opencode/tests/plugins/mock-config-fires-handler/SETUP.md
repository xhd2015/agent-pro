## Steps
1. Write a plugin that logs to a marker file when `session.created` fires.
2. Use mockConfig with `"plugins": ["<plugin.ts>"]`.
3. Verify the marker file was written.

```go
import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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

    configJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_mock","llm_events":[{"type":"message","done":true}],"plugins":["` + pluginPath + `"]}`
    writeMockConfig(t, req, configJSON)

    req.MarkerPath = markerPath
    return nil
}
```
