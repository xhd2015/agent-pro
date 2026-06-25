# fake-opencode Tests

These doc-style tests verify the deterministic fake opencode runner.

```go
import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
    "time"
)


type Request struct {
    RepoRoot string
    TempDir string
    HomeDir string
    FakeOpencode string
    Args []string
    Env []string
    MockConfigPath string
    HookLogPath string
    MarkerPath string
    Operation string
}

type Response struct {
    ExitCode int
    Stdout string
    Stderr string
    HookLog string
    HostConfigExists bool
    Err error
}

func Run(t *testing.T, req *Request) (*Response, error) {
    if req.Operation == "per_event_delay" {
        return runPerEventDelay(t, req)
    }
    args := req.Args
    if len(args) == 0 {
        args = []string{"run", "--format", "json", "--mock-config", req.MockConfigPath, "hello"}
    }
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, req.FakeOpencode, args...)
    cmd.Dir = req.TempDir
    cmd.Env = append(os.Environ(), req.Env...)
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()
    resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
    if data, readErr := os.ReadFile(req.HookLogPath); readErr == nil {
        resp.HookLog = string(data)
    }
    if _, statErr := os.Stat(filepath.Join(req.HomeDir, ".config", "opencode")); statErr == nil {
        resp.HostConfigExists = true
    }
    if err == nil {
        return resp, nil
    }
    if ctx.Err() != nil {
        return resp, ctx.Err()
    }
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        resp.ExitCode = exitErr.ExitCode()
        return resp, nil
    }
    return resp, err
}
```
