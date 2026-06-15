## Preconditions
- The subagent package is at `github.com/xhd2015/agent-pro/agent/subagent`.
- Each grouping node overrides `Run()` with its own test logic.
- The stub `Run()` at root returns an error, ensuring tests start RED until implemented.

## Steps
1. Each grouping node's `SETUP.md` provides a `func Setup` that configures the request.
2. Each grouping node overrides `func Run` with the appropriate test harness.
3. Leaf `ASSERT.md` validates the response.

## Context
- `Req.RoleName`: the sub-agent role name (e.g. "implementer", "designer", "test_role")
- `Req.SessionBase`: passed to `Options.SessionBase`
- `Req.SessionID`: passed to `Options.SessionID`
- `Req.Env`: environment variables for the test
- `Req.Operation`: one of "list_sessions", "show_status", "trace_session", "logf", "resolve_id"
- `Req.LogMessage` / `Req.LogArgs`: for Logf tests
- `Req.PreCreateDirs`: paths to pre-create as session directories (for testing session listing/status)

```go
import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "syscall"
    "testing"
    "time"

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

func Run(t *testing.T, req *Request) (*Response, error) {
    return nil, fmt.Errorf("Run not implemented for %s", req.Operation)
}
```
