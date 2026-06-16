## Preconditions
- `req.AgentRunnerEnv` is `"TEST_AGENT_RUNNER"` (set in parent).
- The `TEST_AGENT_RUNNER` env var is **not** set. Priority 1 is skipped.
- No CODEX_THREAD_ID, no PI_CODING_AGENT → falls through to Priority 4.
- Process names: ppid="bash", pppid="bash" → no agent detected.

## Steps
1. Do not set `TEST_AGENT_RUNNER` env var.
2. Set process names to `["bash", "bash"]` so Priority 4 yields no match.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ProcessNames = []string{"bash", "bash"}
    return nil
}
```
