## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=write and a path.

## Steps
1. Run fake opencode with the mock config.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    writeFile := filepath.Join(req.TempDir, "written.txt")
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_write","llm_events":[{"type":"tool_call","tool":"write","tool_input":{"path":"`+writeFile+`","content":"new file contents"}}]}`)
    return nil
}
```
