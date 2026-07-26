## Preconditions
- Priority 1–3 are not configured (no AgentRunnerEnv, no CODEX_THREAD_ID, no PI_CODING_AGENT).
- The `TestProcessNameFunc` test hook is installed to inject process names.
- `req.ProcessNames` is a sequential list: first call → ppid name, second call → pppid name.
- The hook is reset after each test via `defer`.

## Steps
1. Each leaf sets `req.ProcessNames` to define the fake process tree.
2. `Run` calls `TestExported_autoDetectAgentRunner` which delegates to the hook.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // Priority 4: parent-process detection — ensure no env override, no env detection vars
    req.AgentRunnerEnv = ""
    return nil
}
```
