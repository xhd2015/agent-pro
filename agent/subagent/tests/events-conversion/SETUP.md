## Preconditions
- The `events-conversion` tests verify the events.jsonl pipeline: converting raw runner events to
  `AgentEvent` format via `agent/event/convert/ConvertRawLine`, rendering AgentEvent JSON as
  human-readable strings via `formatEventLine`, and end-to-end trace/status display.
- The root `Run` dispatches based on `req.Operation`: `"convert_raw"`, `"format_event"`, `"trace"`, `"status"`.

## Steps
1. `Setup` at root is a no-op; intermediate nodes configure the request.
2. `Run` dispatches to the appropriate test function based on `req.Operation`.
3. `convert_raw`: calls `convert.ConvertRawLine(req.RawJSON, req.AgentRunner)`, returns AgentEvent JSON.
4. `format_event`: parses `req.AgentEventJSON`, calls `subagent.TestExported_formatEventLine`, returns formatted string.
5. `trace`: pre-creates session dirs with events, calls `subagent.TestExported_traceSession`, captures stdout.
6. `status`: pre-creates session dirs with events, calls `subagent.TestExported_showStatus`, captures stdout.

## Context
- `Req.Operation`: `"convert_raw"`, `"format_event"`, `"trace"`, `"status"`
- `Req.RawJSON`: raw runner-native JSON line for converter tests
- `Req.AgentRunner`: `"opencode"`, `"pi"`, `"codex"`, `"crush"`, or any unknown string
- `Req.AgentEventJSON`: marshaled AgentEvent JSON for formatEventLine tests
- `Req.RoleName`: the sub-agent role name
- `Req.SessionBase`: base path for session directories
- `Req.SessionID`: session ID for trace/status
- `Req.PreCreateDirs`, `Req.PreCreateMeta`, `Req.PreCreateEvents`, `Req.PreCreatePID`: session scaffolding

```go
import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "sync"
    "testing"
    "time"

    types "github.com/xhd2015/agent-pro/agent/event/types"

    "github.com/xhd2015/agent-pro/agent/event/convert"
    "github.com/xhd2015/agent-pro/agent/subagent"
)

var captureIOMu sync.Mutex

func captureStdoutStderr(fn func() error) (stdout, stderr string, err error) {
    captureIOMu.Lock()
    defer captureIOMu.Unlock()
    oldOut, oldErr := os.Stdout, os.Stderr
    rOut, wOut, e := os.Pipe()
    if e != nil {
        return "", "", e
    }
    rErr, wErr, e := os.Pipe()
    if e != nil {
        wOut.Close()
        rOut.Close()
        return "", "", e
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
    _ = runConvertRaw
    _ = runFormatEvent
    _ = runTrace
    _ = runStatus
    return nil
}

func runConvertRaw(t *testing.T, req *Request) (*Response, error) {
    events, err := convert.ConvertRawLine([]byte(req.RawJSON), req.AgentRunner)
    if err != nil {
        return nil, err
    }
    data, _ := json.Marshal(events)
    return &Response{Stdout: string(data)}, nil
}

func runFormatEvent(t *testing.T, req *Request) (*Response, error) {
    output := subagent.TestExported_formatEventLine(req.AgentEventJSON)
    return &Response{Stdout: output}, nil
}

func runTrace(t *testing.T, req *Request) (*Response, error) {
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

    var err error
    stdout, _, _ := captureStdoutStderr(func() error {
        err = subagent.TestExported_traceSession(subagent.Config{
            RoleName: roleName,
            HomeDir:  homeDir,
        }, subagent.Options{
            CatchUp:     true,
            SessionID:   req.SessionID,
            SessionBase: req.SessionBase,
        })
        return nil
    })

    return &Response{Stdout: stdout, Err: err}, nil
}

func runStatus(t *testing.T, req *Request) (*Response, error) {
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

    var err error
    stdout, stderr, _ := captureStdoutStderr(func() error {
        err = subagent.TestExported_showStatus(subagent.Config{
            RoleName: roleName,
            HomeDir:  homeDir,
        }, subagent.Options{
            Status:      true,
            SessionID:   req.SessionID,
            SessionBase: req.SessionBase,
        })
        return nil
    })

    return &Response{Stdout: stdout, Stderr: stderr, Err: err}, nil
}
```
