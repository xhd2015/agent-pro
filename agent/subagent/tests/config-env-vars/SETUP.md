## Preconditions
- Tests for `Config.SessionEnvVar`, `Config.SessionMetaField`, and `Config.DebugSessionEnv`.
- When set, these override the default env var / meta field / debug env names.
- When empty, defaults are used.

## Steps
1. Each sub-grouping node configures the relevant Config field(s).
2. Each leaf sets the custom or default value and verifies behavior via the public API.

```go
import (
    "bytes"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "testing"
)

func Run(t *testing.T, req *Request) (*Response, error) {
    for _, e := range req.Env {
        parts := splitEnv(e)
        if len(parts) == 2 {
            os.Setenv(parts[0], parts[1])
        }
    }

    homeDir := req.HomeDir
    if homeDir == "" {
        homeDir = t.TempDir()
    }
    os.Setenv("HOME", homeDir)

    os.Unsetenv("DOCTEST_DEBUG_SESSION_HOME")
    os.Unsetenv("AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME")
    os.Unsetenv("AGENT_PRO_SUBAGENT_" + strings.ToUpper(req.RoleName) + "_SESSION_ID")
    os.Unsetenv("CODEX_THREAD_ID")

    for _, e := range req.Env {
        parts := splitEnv(e)
        if len(parts) == 2 {
            os.Setenv(parts[0], parts[1])
        }
    }

    for _, dir := range req.PreCreateDirs {
        os.MkdirAll(dir, 0755)
    }
    for dir, content := range req.PreCreateMeta {
        os.WriteFile(filepath.Join(dir, "meta.json"), []byte(content), 0644)
    }
    for dir, content := range req.PreCreateEvents {
        os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(content), 0644)
    }
    if req.PreCreatePID {
        for _, dir := range req.PreCreateDirs {
            os.WriteFile(filepath.Join(dir, "pid"), []byte(strconv.Itoa(os.Getpid())), 0644)
        }
    }

    oldOut := os.Stdout
    rOut, wOut, _ := os.Pipe()
    os.Stdout = wOut

    oldErr := os.Stderr
    rErr, wErr, _ := os.Pipe()
    os.Stderr = wErr

    roleName := req.RoleName
    if roleName == "" {
        roleName = "testrole"
    }

    cfg := subagent.Config{
        RoleName:         roleName,
        SessionEnvVar:    req.SessionEnvVar,
        SessionMetaField: req.SessionMetaField,
        DebugSessionEnv:  req.DebugSessionEnv,
    }

    var runErr error
    if req.Status {
        runErr = subagent.Run(cfg, subagent.Options{
            Status:      true,
            SessionID:   req.SessionID,
            SessionBase: req.SessionBase,
        })
    } else if req.ListSessions {
        runErr = subagent.Run(cfg, subagent.Options{
            ListSessions: true,
            SessionBase:  req.SessionBase,
        })
    } else if req.CatchUp {
        runErr = subagent.Run(cfg, subagent.Options{
            CatchUp:     true,
            SessionID:   req.SessionID,
            SessionBase: req.SessionBase,
        })
    } else {
        runErr = fmt.Errorf("no operation mode set")
    }

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

func splitEnv(e string) []string {
    for i := 0; i < len(e); i++ {
        if e[i] == '=' {
            return []string{e[:i], e[i+1:]}
        }
    }
    return []string{e}
}
```
