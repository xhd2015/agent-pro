## Preconditions
- `bash`, `go`, `node`, and `bun` are available in PATH (individual tests skip if not present).
- No code under test from this project — these are pure platform behavior demonstrations.

## Steps
1. Create a temporary directory to serve as the fake `HOME`.
2. Run the target runtime with `HOME` set to the temporary directory.
3. Capture the runtime's reported home directory.
4. Assert it matches the temporary `HOME` directory.

```go
import (
    "bytes"
    "context"
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
    TmpHome string // temporary directory used as fake HOME
    Cmd     string // path to the runtime binary (bash, go, node, bun)
    Args    []string
    WorkDir string // if set, the command runs from this directory
    Env     []string // additional env vars (HOME is always appended)
}

type Response struct {
    ExitCode int
    Stdout   string
    Stderr   string
    Err      error
}

func Setup(t *testing.T, req *Request) error {
    req.TmpHome = filepath.Join(t.TempDir(), "fake-home")
    if err := os.MkdirAll(req.TmpHome, 0755); err != nil {
        return fmt.Errorf("create fake HOME: %w", err)
    }
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, req.Cmd, req.Args...)
    if req.WorkDir != "" {
        cmd.Dir = req.WorkDir
    }
    cmd.Env = append(os.Environ(), req.Env...)
    cmd.Env = append(cmd.Env, "HOME="+req.TmpHome)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
    if err != nil {
        if ctx.Err() != nil {
            return resp, ctx.Err()
        }
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            resp.ExitCode = exitErr.ExitCode()
        }
    }
    return resp, nil
}

// assertHomeDir is a helper to check that the runtime's reported home
// matches the temporary HOME directory.
func assertHomeDir(t *testing.T, req *Request, resp *Response, err error) {
    t.Helper()
    if err != nil && resp.ExitCode == 0 {
        t.Fatalf("run failed: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
    }
    got := strings.TrimSpace(resp.Stdout)
    if got != req.TmpHome {
        t.Fatalf("expected home %q, got %q", req.TmpHome, got)
    }
}
```
