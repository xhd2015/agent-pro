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

type Request struct {
    Env            []string
    AgentRunnerEnv string
    ProcessNames   []string
}

type Response struct {
    Runner   string
    Detected bool
}

func Setup(t *testing.T, req *Request) error {
    _ = req.AgentRunnerEnv
    _ = req.ProcessNames
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    // Clean detection env vars that may leak from the host environment
    os.Unsetenv("CODEX_THREAD_ID")
    os.Unsetenv("PI_CODING_AGENT")

    for _, e := range req.Env {
        parts := splitEnv(e)
        if len(parts) == 2 {
            os.Setenv(parts[0], parts[1])
        }
    }

    if len(req.ProcessNames) > 0 {
        callIdx := 0
        subagent.TestProcessNameFunc = func(pid int) string {
            if callIdx < len(req.ProcessNames) {
                name := req.ProcessNames[callIdx]
                callIdx++
                return name
            }
            return ""
        }
        defer func() { subagent.TestProcessNameFunc = nil }()
    }

    runner, detected := subagent.TestExported_autoDetectAgentRunner(subagent.Config{
        AgentRunnerEnv: req.AgentRunnerEnv,
    })

    return &Response{Runner: runner, Detected: detected}, nil
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
