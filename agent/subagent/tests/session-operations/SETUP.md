## Preconditions
- `session-operations` tests verify that `listSessions`, `showStatus`, and `traceSession` work correctly with the configurable base directory.

## Steps
1. Create session directories from `req.PreCreateDirs` with metadata and events from `req.PreCreateMeta` and `req.PreCreateEvents`.
2. If `req.PreCreatePID` is true, write a pid file with the current process PID.
3. Call `subagent.Run()` with the operation flag (`ListSessions`, `Status`, or `CatchUp`).
4. Capture stdout and stderr, return as response.

```go
import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "testing"
    "time"
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
        req.HomeDir = homeDir
    }
    os.Setenv("HOME", homeDir)

    os.Unsetenv("DOCTEST_DEBUG_SESSION_HOME")
    os.Unsetenv("AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME")
    for _, e := range req.Env {
        parts := splitEnv(e)
        if len(parts) == 2 && parts[0] == "AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME" {
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

    var runErr error
    if req.ListSessions {
        runErr = subagent.Run(subagent.Config{
            RoleName: roleName,
        }, subagent.Options{
            ListSessions: true,
            SessionBase:  req.SessionBase,
        })
    } else if req.Status {
        runErr = subagent.Run(subagent.Config{
            RoleName: roleName,
        }, subagent.Options{
            Status:      true,
            SessionID:   req.SessionID,
            SessionBase: req.SessionBase,
        })
    } else if req.CatchUp {
        runErr = subagent.Run(subagent.Config{
            RoleName: roleName,
        }, subagent.Options{
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
