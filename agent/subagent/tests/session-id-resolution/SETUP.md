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
- Env isolation uses request-scoped `Config.EnvLookup` (no process Setenv).

## Steps
1. Build EnvLookup that force-unsets role session env and `CODEX_THREAD_ID`, then applies `req.Env`.
2. Build `subagent.Config` with `AutoGenerateSessionID` from the request.
3. Run `subagent.Run` with `Status: true` and `req.SessionID`.
4. Return captured stdout, stderr, and error.

```go
import (
    "bytes"
    "context"
    "fmt"
    "os"
    "strings"
    "sync"
    "testing"

    "github.com/xhd2015/agent-pro/agent/subagent"
)

var captureIOMu sync.Mutex

func captureStdoutStderr(fn func() error) (stdout, stderr string, err error) {
    captureIOMu.Lock()
    defer captureIOMu.Unlock()

    oldOut, oldErr := os.Stdout, os.Stderr
    rOut, wOut, e := os.Pipe()
    if e != nil {
        return "", "", fmt.Errorf("stdout pipe: %w", e)
    }
    rErr, wErr, e := os.Pipe()
    if e != nil {
        wOut.Close()
        rOut.Close()
        return "", "", fmt.Errorf("stderr pipe: %w", e)
    }
    os.Stdout, os.Stderr = wOut, wErr
    err = fn()
    wOut.Close()
    wErr.Close()
    os.Stdout, os.Stderr = oldOut, oldErr
    var bufOut, bufErr bytes.Buffer
    bufOut.ReadFrom(rOut)
    bufErr.ReadFrom(rErr)
    return bufOut.String(), bufErr.String(), err
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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

func envLookup(env []string, forceUnset ...string) func(string) (string, bool) {
    overlay := map[string]*string{}
    for _, k := range forceUnset {
        overlay[k] = nil
    }
    for _, e := range env {
        parts := splitEnv(e)
        if len(parts) != 2 {
            continue
        }
        v := parts[1]
        overlay[parts[0]] = &v
    }
    return func(key string) (string, bool) {
        p, ok := overlay[key]
        if !ok {
            return "", false
        }
        if p == nil {
            return "", true
        }
        return *p, true
    }
}

func runSessionIDResolution(t *testing.T, req *Request) (*Response, error) {
    roleName := req.RoleName
    if roleName == "" {
        roleName = "testrole"
    }

    roleEnv := "AGENT_PRO_SUBAGENT_" + strings.ToUpper(roleName) + "_SESSION_ID"

    var runErr error
    stdout, stderr, _ := captureStdoutStderr(func() error {
        runErr = subagent.Run(context.Background(), subagent.Config{
            RoleName:              roleName,
            AutoGenerateSessionID: req.AutoGenerateSessionID,
            EnvLookup:             envLookup(req.Env, roleEnv, "CODEX_THREAD_ID"),
        }, subagent.Options{
            Status:    true,
            SessionID: req.SessionID,
        })
        return nil
    })

    return &Response{Stdout: stdout, Stderr: stderr, Err: runErr}, nil
}
```
