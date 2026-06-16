## Preconditions
- The subagent package is at `github.com/xhd2015/agent-pro/agent/subagent`.
- The root `Run` dispatches to operation-specific handlers based on `req.Operation`.
- Each grouping node's `SETUP.md` provides a `func Setup` that configures the request.
- Leaf `ASSERT.md` validates the response.

## Steps
1. `Setup` at root is a no-op; intermediate nodes configure `req.Operation` and request fields.
2. `Run` dispatches: `"logf"` → runLogf, `"session_base"` → runSessionBase, `"session_id_resolution"` → runSessionIDResolution, default → runSubagentOp.
3. `splitEnv` is a shared helper for parsing KEY=VALUE pairs.

## Context
- `Req.RoleName`: the sub-agent role name (e.g. "implementer", "designer", "test_role")
- `Req.SessionBase`: passed to `Options.SessionBase`
- `Req.SessionID`: passed to `Options.SessionID`
- `Req.Env`: environment variables for the test
- `Req.Operation`: one of "debug_session_env", "session_env_var", "session_meta_field", "list_sessions", "show_status", "trace_session", "logf", "session_base", "session_id_resolution"
- `Req.LogMessage` / `Req.LogArgs`: for Logf tests
- `Req.PreCreateDirs`: paths to pre-create as session directories (for testing session listing/status)

```go
import (
    "bytes"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "testing"

    "github.com/xhd2015/agent-pro/agent/subagent"
)

type Request struct {
    RoleName    string
    SessionBase string
    SessionID   string
    Env         []string

    Operation      string
    ListSessions   bool
    Status         bool
    CatchUp        bool

    LogMessage string
    LogArgs    []any

    PreCreateDirs    []string
    PreCreateMeta    map[string]string
    PreCreateEvents  map[string]string
    PreCreatePID     bool

    HomeDir string

    SessionEnvVar   string
    SessionMetaField string
    DebugSessionEnv  string
}

type Response struct {
    Stdout string
    Stderr string
    Err    error
}

func Setup(t *testing.T, req *Request) error {
    _ = runLogf
    _ = runSessionBase
    _ = runSessionIDResolution
    _ = runSubagentOp
    _ = splitEnv
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    switch req.Operation {
    case "logf":
        return runLogf(t, req)
    case "session_base":
        return runSessionBase(t, req)
    case "session_id_resolution":
        return runSessionIDResolution(t, req)
    default:
        return runSubagentOp(t, req)
    }
}

func splitEnv(e string) []string {
    for i := 0; i < len(e); i++ {
        if e[i] == '=' {
            return []string{e[:i], e[i+1:]}
        }
    }
    return []string{e}
}

func runLogf(t *testing.T, req *Request) (*Response, error) {
    message := req.LogMessage

    old := os.Stdout
    r, w, err := os.Pipe()
    if err != nil {
        return nil, fmt.Errorf("create pipe: %w", err)
    }
    os.Stdout = w

    subagent.Logf("%s", fmt.Sprintf(message, req.LogArgs...))

    w.Close()
    os.Stdout = old

    var buf bytes.Buffer
    if _, readErr := buf.ReadFrom(r); readErr != nil {
        return nil, fmt.Errorf("read pipe: %w", readErr)
    }

    return &Response{Stdout: buf.String()}, nil
}

func runSessionBase(t *testing.T, req *Request) (*Response, error) {
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

    runErr := subagent.Run(context.Background(), subagent.Config{
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

func runSessionIDResolution(t *testing.T, req *Request) (*Response, error) {
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
        RoleName: req.RoleName,
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

func runSubagentOp(t *testing.T, req *Request) (*Response, error) {
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
    os.Unsetenv("AGENT_PRO_SUBAGENT_"+strings.ToUpper(req.RoleName)+"_SESSION_ID")
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
        runErr = subagent.Run(context.Background(), cfg, subagent.Options{
            Status:      true,
            SessionID:   req.SessionID,
            SessionBase: req.SessionBase,
        })
    } else if req.ListSessions {
        runErr = subagent.Run(context.Background(), cfg, subagent.Options{
            ListSessions: true,
            SessionBase:  req.SessionBase,
        })
    } else if req.CatchUp {
        runErr = subagent.Run(context.Background(), cfg, subagent.Options{
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
```
