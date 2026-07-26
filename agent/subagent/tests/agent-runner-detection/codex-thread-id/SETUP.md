## Preconditions
- Priority 1 (AgentRunnerEnv) is not configured.
- The `CODEX_THREAD_ID` env var is set, triggering Priority 2 detection.

## Steps
1. Ensure `req.AgentRunnerEnv` is not set (blank).
2. Each leaf sets `CODEX_THREAD_ID` in the environment.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // Priority 2: CODEX_THREAD_ID detection — ensure no env override
    req.AgentRunnerEnv = ""
    return nil
}
```
