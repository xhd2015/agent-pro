# Scenario

**Feature**: subagent library session management, env vars, and Logf behavior

```
# harness dispatches by Operation, builds Config/Options, captures I/O
test Request -> Run(operation) -> subagent.Run -> stdout/stderr

# session ID resolution checks flag, env var, CODEX_THREAD_ID, then policy
resolveSessionID(Config, flag, prompt) -> sessionIDSources | error | generated
```

## Preconditions
- The subagent package is at `github.com/xhd2015/agent-pro/agent/subagent`.
- The root `Run` dispatches to operation-specific handlers based on `req.Operation`.
- Each grouping node's `SETUP.md` provides a `func Setup` that configures the request.
- Leaf `ASSERT.md` validates the response.

## Steps
1. `Setup` at root is a no-op; intermediate nodes configure `req.Operation` and request fields.
2. `Run` dispatches: `"logf"` → runLogf, `"session_base"` → runSessionBase, `"session_id_resolution"` → runSessionIDResolution, default → runSubagentOp.
3. `splitEnv` / `envLookup` are shared helpers; env isolation uses request-scoped `Config.EnvLookup` and `Config.HomeDir` (no process Setenv).

## Context
- `Req.RoleName`: the sub-agent role name (e.g. "implementer", "designer", "test_role")
- `Req.SessionBase`: passed to `Options.SessionBase`
- `Req.SessionID`: passed to `Options.SessionID`
- `Req.Env`: environment variables for the test (via Config.EnvLookup, not os.Setenv)
- `Req.Operation`: one of "debug_session_env", "session_env_var", "session_meta_field", "list_sessions", "show_status", "trace_session", "logf", "session_base", "session_id_resolution"
- `Req.LogMessage` / `Req.LogArgs`: for Logf tests
- `Req.PreCreateDirs`: paths to pre-create as session directories (for testing session listing/status)
- `Req.AutoGenerateSessionID`: passed to `Config.AutoGenerateSessionID` for session ID policy

```go
import (
    "bytes"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    "testing"

    "github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = runLogf
    _ = runSessionBase
    _ = runSessionIDResolution
    _ = runSubagentOp
    _ = splitEnv
    _ = envLookup
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

// envLookup builds a request-scoped EnvLookup: force-unset keys return ("", true);
// req.Env KEY=VALUE pairs return (value, true); other keys fall through to process env.
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

func runLogf(t *testing.T, req *Request) (*Response, error) {
    message := req.LogMessage
    stdout, _, err := captureStdoutStderr(func() error {
        subagent.Logf("%s", fmt.Sprintf(message, req.LogArgs...))
        return nil
    })
    if err != nil {
        return nil, err
    }
    return &Response{Stdout: stdout}, nil
}

// captureStdoutStderr serializes os.Stdout/os.Stderr swaps so parallel leaves
// do not steal each other's pipes (process-global streams).
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

func runSessionBase(t *testing.T, req *Request) (*Response, error) {
    homeDir := req.HomeDir
    if homeDir == "" {
        homeDir = t.TempDir()
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

    var runErr error
    stdout, _, _ := captureStdoutStderr(func() error {
        runErr = subagent.Run(context.Background(), subagent.Config{
            RoleName:  req.RoleName,
            HomeDir:   homeDir,
            EnvLookup: envLookup(req.Env, "AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME"),
        }, subagent.Options{
            ListSessions: true,
            SessionBase:  req.SessionBase,
        })
        return nil
    })

    return &Response{Stdout: stdout, Err: runErr}, nil
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

func runSubagentOp(t *testing.T, req *Request) (*Response, error) {
    homeDir := req.HomeDir
    if homeDir == "" {
        homeDir = t.TempDir()
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

    roleName := req.RoleName
    if roleName == "" {
        roleName = "testrole"
    }

    roleEnv := "AGENT_PRO_SUBAGENT_" + strings.ToUpper(roleName) + "_SESSION_ID"
    cfg := subagent.Config{
        RoleName:         roleName,
        SessionEnvVar:    req.SessionEnvVar,
        SessionMetaField: req.SessionMetaField,
        DebugSessionEnv:  req.DebugSessionEnv,
        HomeDir:          homeDir,
        EnvLookup: envLookup(req.Env,
            "DOCTEST_DEBUG_SESSION_HOME",
            "AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME",
            roleEnv,
            "CODEX_THREAD_ID",
        ),
    }

    var runErr error
    stdout, stderr, _ := captureStdoutStderr(func() error {
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
        return nil
    })

    return &Response{Stdout: stdout, Stderr: stderr, Err: runErr}, nil
}
```
