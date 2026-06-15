## Preconditions
- `session-base` tests verify that `sessionsBase` (via the public `Run()` API with `ListSessions`) respects `Options.SessionBase`, the default `~/.agent-pro/subagent/<role>/sessions/`, and the `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME` env var override.

## Steps
1. Create session directories under the target base directory.
2. Set `req.HomeDir` to a temp directory so `~` resolves predictably.
3. Set env vars from `req.Env`.
4. Call `subagent.Run()` with `ListSessions: true`.
5. Capture stdout and return.

```go
import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "testing"
    "time"
)

func Setup(t *testing.T, req *Request) error {
    req.Operation = "session_base"
    req.RoleName = "testrole"
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    homeDir := req.HomeDir
    if homeDir == "" {
        homeDir = t.TempDir()
    }
    os.Setenv("HOME", homeDir)

    for _, e := range req.Env {
        parts := splitEnv(e)
        if len(parts) == 2 {
            os.Setenv(parts[0], parts[1])
        }
    }
    os.Unsetenv("AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME")
    for _, e := range req.Env {
        parts := splitEnv(e)
        if len(parts) == 2 && parts[0] == "AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME" {
            os.Setenv(parts[0], parts[1])
        }
    }

    for _, dir := range req.PreCreateDirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return nil, fmt.Errorf("create dir %s: %w", dir, err)
        }
    }
    for dir, content := range req.PreCreateMeta {
        if err := os.WriteFile(filepath.Join(dir, "meta.json"), []byte(content), 0644); err != nil {
            return nil, fmt.Errorf("write meta %s: %w", dir, err)
        }
    }

    old := os.Stdout
    r, w, err := os.Pipe()
    if err != nil {
        return nil, fmt.Errorf("create pipe: %w", err)
    }
    os.Stdout = w

    runErr := subagent.Run(subagent.Config{
        RoleName: req.RoleName,
    }, subagent.Options{
        ListSessions: true,
        SessionBase:  req.SessionBase,
    })

    w.Close()
    os.Stdout = old

    var buf bytes.Buffer
    buf.ReadFrom(r)

    return &Response{Stdout: buf.String(), Err: runErr}, nil
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
