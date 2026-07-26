## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=read and a path to an existing file.

## Steps
1. Create a file to read.
2. Run fake opencode with the mock config.

```go
import (
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    readFile := filepath.Join(req.TempDir, "data.txt")
    writeFile(t, readFile, "readable content")
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_read","llm_events":[{"type":"tool_call","tool":"read","tool_input":{"path":"`+readFile+`"}}]}`)
    return nil
}
```
