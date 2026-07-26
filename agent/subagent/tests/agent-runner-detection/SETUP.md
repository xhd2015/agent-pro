## Preconditions
- The `autoDetectAgentRunner` function in `github.com/xhd2015/agent-pro/agent/subagent` detects the hosting agent runner via four priority levels.
- `subagent.TestExported_autoDetectAgentRunner` wraps the unexported function.
- `subagent.TestProcessNameFunc` is the test hook for injecting process names.

## Steps
1. `Setup` chain configures `Request` with environment variables, process names, and `AgentRunnerEnv`.
2. `Run` sets env vars, installs the process-name hook (when `req.ProcessNames` is set), and calls `TestExported_autoDetectAgentRunner`.
3. `Assert` validates `resp.Runner` and `resp.Detected`.

## Context
- `req.Env`: environment variables to set, in `KEY=VALUE` format.
- `req.AgentRunnerEnv`: value for `Config.AgentRunnerEnv` (the env var name whose value is checked at Priority 1).
- `req.ProcessNames`: sequential list of process names returned by the hook. First call → ppid name, second call → pppid name.

```go
import (
    "os"
    "strings"
    "testing"

    "github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = req.AgentRunnerEnv
    _ = req.ProcessNames
    return nil
}

func splitEnv(e string) []string {
    for i := 0; i < len(e); i++ {
        if e[i] == '=' {
            return []string{e[:i], e[i+1:]}
        }
    }
    return []string{e}
}
```
