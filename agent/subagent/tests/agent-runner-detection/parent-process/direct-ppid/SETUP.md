## Preconditions
- The immediate parent process (ppid) matches a known agent runner.
- Only the ppid is checked; no grandparent walk is needed.
- Covers all five known agents: pi, opencode, codex, crush, grok.

## Steps
1. Each leaf sets `req.ProcessNames` with a single entry: the agent process name.
2. Only one call to the process-name hook is expected (ppid match returns early).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    // Direct ppid match — no grandparent walk needed
    req.AgentRunnerEnv = ""
    return nil
}
```
