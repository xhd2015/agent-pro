## Preconditions
- `req.AgentRunnerEnv` is `"TEST_AGENT_RUNNER"` (set in parent).
- The `TEST_AGENT_RUNNER` env var is set to `"pi"`.

## Steps
1. Set env: `TEST_AGENT_RUNNER=pi`.
2. Call `TestExported_autoDetectAgentRunner`. Priority 1 matches, returns immediately.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = []string{"TEST_AGENT_RUNNER=pi"}
    return nil
}
```
