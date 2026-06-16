## Preconditions
- Priority 1 (AgentRunnerEnv) and Priority 2 (CODEX_THREAD_ID) are not configured.
- The `PI_CODING_AGENT` env var is set, triggering Priority 3 detection.

## Steps
1. Ensure `req.AgentRunnerEnv` is not set (blank).
2. Each leaf sets `PI_CODING_AGENT` in the environment.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    // Priority 3: PI_CODING_AGENT detection — ensure no env override or CODEX_THREAD_ID
    req.AgentRunnerEnv = ""
    return nil
}
```
