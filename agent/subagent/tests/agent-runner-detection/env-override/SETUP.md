## Preconditions
- Priority 1: `Config.AgentRunnerEnv` is set to `"TEST_AGENT_RUNNER"`.
- The env var `TEST_AGENT_RUNNER` may or may not be set, depending on the leaf.

## Steps
1. Set `req.AgentRunnerEnv = "TEST_AGENT_RUNNER"`.
2. Each leaf configures whether the `TEST_AGENT_RUNNER` env var is set and to what value.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.AgentRunnerEnv = "TEST_AGENT_RUNNER"
    return nil
}
```
