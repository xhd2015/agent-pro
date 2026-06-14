## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=grep and a pattern to search.

## Steps
1. Create a file containing a matching pattern.
2. Run fake Codex with the mock config.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    grepFile := filepath.Join(req.TempDir, "target.go")
    writeFile(t, grepFile, "package main\nfunc main() {\n\tfmt.Println(\"needle\")\n}\n")
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"tool_call","tool":"grep","tool_input":{"pattern":"needle","path":"`+req.TempDir+`"}}]}`)
    return nil
}
```
