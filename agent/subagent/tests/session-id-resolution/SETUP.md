# Scenario

**Feature**: session ID resolution via flag, env var, CODEX_THREAD_ID, and policy

```
# showStatus triggers resolveSessionID before session lookup
Run(Status) -> resolveSessionID(Config, --session-id, prompt) -> findSession

# priority: flag > AGENT_PRO_SUBAGENT_<ROLE>_SESSION_ID > CODEX_THREAD_ID > policy
--session-id -> env var -> CODEX_THREAD_ID -> Require error | AutoGenerate gen_*
```

## Preconditions
- Tests call `subagent.Run` with `Status: true` and capture stdout/stderr.
- `Config.AutoGenerateSessionID` controls behavior when all ID sources miss.

## Steps
1. Unset role session env var and `CODEX_THREAD_ID`, then apply `req.Env`.
2. Build `subagent.Config` with `AutoGenerateSessionID` from the request.
3. Run `subagent.Run` with `Status: true` and `req.SessionID`.
4. Return captured stdout, stderr, and error.

```go
import (
    "bytes"
    "context"
    "os"
    "strings"
    "testing"

    "github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, req *Request) error {
    if req.RoleName == "" {
        req.RoleName = "testrole"
    }
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

func runSessionIDResolution(t *testing.T, req *Request) (*Response, error) {
    roleName := req.RoleName
    if roleName == "" {
        roleName = "testrole"
    }

    os.Unsetenv("AGENT_PRO_SUBAGENT_" + strings.ToUpper(roleName) + "_SESSION_ID")
    os.Unsetenv("CODEX_THREAD_ID")

    for _, e := range req.Env {
        parts := splitEnv(e)
        if len(parts) == 2 {
            os.Setenv(parts[0], parts[1])
        }
    }

    oldOut := os.Stdout
    rOut, wOut, _ := os.Pipe()
    os.Stdout = wOut

    oldErr := os.Stderr
    rErr, wErr, _ := os.Pipe()
    os.Stderr = wErr

    runErr := subagent.Run(context.Background(), subagent.Config{
        RoleName:              roleName,
        AutoGenerateSessionID: req.AutoGenerateSessionID,
    }, subagent.Options{
        Status:    true,
        SessionID: req.SessionID,
    })

    wOut.Close()
    wErr.Close()
    os.Stdout = oldOut
    os.Stderr = oldErr

    var bufOut bytes.Buffer
    bufOut.ReadFrom(rOut)
    var bufErr bytes.Buffer
    bufErr.ReadFrom(rErr)

    return &Response{Stdout: bufOut.String(), Stderr: bufErr.String(), Err: runErr}, nil
}
```