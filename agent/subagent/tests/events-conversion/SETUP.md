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
    "testing"
    "time"

    types "github.com/xhd2015/agent-pro/agent/event/types"

    "github.com/xhd2015/agent-pro/agent/event/convert"
    "github.com/xhd2015/agent-pro/agent/subagent"
)

type Request struct {
    Operation   string // "convert_raw" | "format_event" | "trace" | "status"

    // For convert_raw
    RawJSON     string
    AgentRunner string

    // For format_event
    AgentEventJSON string

    // For trace / status
    RoleName        string
    SessionBase     string
    SessionID       string
    HomeDir         string
    PreCreateDirs   []string
    PreCreateMeta   map[string]string
    PreCreateEvents map[string]string
    PreCreatePID    bool
}

type Response struct {
    Stdout string
    Stderr string
    Err    error
}

func Setup(t *testing.T, req *Request) error {
    _ = runConvertRaw
    _ = runFormatEvent
    _ = runTrace
    _ = runStatus
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    switch req.Operation {
    case "convert_raw":
        return runConvertRaw(t, req)
    case "format_event":
        return runFormatEvent(t, req)
    case "trace":
        return runTrace(t, req)
    case "status":
        return runStatus(t, req)
    default:
        return nil, fmt.Errorf("unknown operation: %s", req.Operation)
    }
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
    os.Setenv("HOME", homeDir)

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

    oldOut := os.Stdout
    rOut, wOut, _ := os.Pipe()
    os.Stdout = wOut

    err := subagent.TestExported_traceSession(subagent.Config{
        RoleName: roleName,
    }, subagent.Options{
        CatchUp:     true,
        SessionID:   req.SessionID,
        SessionBase: req.SessionBase,
    })

    wOut.Close()
    os.Stdout = oldOut

    var bufOut bytes.Buffer
    bufOut.ReadFrom(rOut)

    return &Response{Stdout: bufOut.String(), Err: err}, nil
}

func runStatus(t *testing.T, req *Request) (*Response, error) {
    homeDir := req.HomeDir
    if homeDir == "" {
        homeDir = t.TempDir()
    }
    os.Setenv("HOME", homeDir)

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

    oldOut := os.Stdout
    rOut, wOut, _ := os.Pipe()
    os.Stdout = wOut

    oldErr := os.Stderr
    rErr, wErr, _ := os.Pipe()
    os.Stderr = wErr

    err := subagent.TestExported_showStatus(subagent.Config{
        RoleName: roleName,
    }, subagent.Options{
        Status:      true,
        SessionID:   req.SessionID,
        SessionBase: req.SessionBase,
    })

    wOut.Close()
    wErr.Close()
    os.Stdout = oldOut
    os.Stderr = oldErr

    var bufOut bytes.Buffer
    bufOut.ReadFrom(rOut)
    var bufErr bytes.Buffer
    bufErr.ReadFrom(rErr)

    return &Response{Stdout: bufOut.String(), Stderr: bufErr.String(), Err: err}, nil
}
```
