## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=write and a path.
- No mock output, so the file is actually written.

## Steps
1. Run fake Codex with the mock config.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    writeFile := filepath.Join(req.TempDir, "output.txt")
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"tool_call","tool":"write","tool_input":{"path":"`+writeFile+`","content":"generated content"}}]}`)
    return nil
}
```
