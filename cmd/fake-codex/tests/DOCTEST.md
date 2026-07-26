# fake-codex Mock Config Tests

These doc-style tests verify `cmd/fake-codex exec --json --mock-config`.

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
doctest test ./cmd/fake-codex/tests                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/fake-codex/tests
doctest test --label-all ./cmd/fake-codex/tests
```


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
	"github.com/xhd2015/doctest/session"
)


type Request struct {
    RepoRoot string
    TempDir string
    FakeCodex string
    Args []string
    Env []string
    MockConfigPath string
    LegacyScriptPath string
    HookLogPath string
    MarkerPath string
    Start time.Time
}

type Response struct {
    ExitCode int
    Stdout string
    Stderr string
    HookLog string
    MarkerLog string
    Err error
    Duration time.Duration
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
    args := req.Args
    if len(args) == 0 {
        args = []string{"exec", "--json", "--mock-config", req.MockConfigPath, "hello"}
    }
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, req.FakeCodex, args...)
    cmd.Dir = req.TempDir
    cmd.Env = append(os.Environ(), req.Env...)
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    req.Start = time.Now()
    err := cmd.Run()
    duration := time.Since(req.Start)
    resp := &Response{
        Stdout: stdout.String(),
        Stderr: stderr.String(),
        Err: err,
        Duration: duration,
    }
    if data, readErr := os.ReadFile(req.HookLogPath); readErr == nil {
        resp.HookLog = string(data)
    }
    if data, readErr := os.ReadFile(req.MarkerPath); readErr == nil {
        resp.MarkerLog = string(data)
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
