## Preconditions
- The mock config contains a read tool_call AgentEvent pointing to a non-existent file.

## Steps
1. Use a path that does not exist on disk.
2. Run fake-opencode and verify an error is surfaced.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_read_nf","llm_events":[{"type":"tool_call","tool":"read","tool_input":{"path":"/nonexistent/path/to/file_that_does_not_exist.txt"}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
