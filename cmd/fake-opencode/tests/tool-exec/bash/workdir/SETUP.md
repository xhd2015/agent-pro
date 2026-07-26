## Preconditions
- The mock config contains a bash tool_call AgentEvent with a `workdir` in the input.

## Steps
1. Create a known subdirectory in the temp dir.
2. Write a mock config where bash runs `pwd` with workdir set to that subdirectory.
3. Run fake-opencode and verify pwd output matches the workdir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    workDir := createTestFile(t, req, "subdir/placeholder.txt", "marker")
    _ = workDir
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash_wd","llm_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"pwd","workdir":"` + req.TempDir + `/subdir"}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
