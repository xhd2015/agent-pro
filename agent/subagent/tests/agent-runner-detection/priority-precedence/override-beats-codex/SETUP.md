## Preconditions
- `AgentRunnerEnv` is `"TEST_AGENT_RUNNER"` (Priority 1 configured).
- `TEST_AGENT_RUNNER=pi` is set in env → Priority 1 returns "pi".
- `CODEX_THREAD_ID=abc123` is also set, but Priority 1 fires first.

## Steps
1. Set `TEST_AGENT_RUNNER=pi` and `CODEX_THREAD_ID=abc123`.
2. Priority 1 matches → returns `"pi"`, `true` (not "codex").

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.AgentRunnerEnv = "TEST_AGENT_RUNNER"
    req.Env = []string{
        "TEST_AGENT_RUNNER=pi",
        "CODEX_THREAD_ID=abc123",
    }
    return nil
}
```
