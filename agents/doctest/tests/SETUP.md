## Preconditions
- The repository root contains the doctest command at `./agents/doctest`.
- The tests are executed by the doc-style test runner from this test tree.
- Each leaf sets the doctest arguments it wants to execute.

## Steps
1. Resolve the repository root from `DOCTEST_ROOT`.
2. Execute the binary given by `req.Bin`.
3. Capture stdout, stderr, exit code, and the raw execution error.

## Context
- These are real integration tests, not mocked unit tests.
- Agent tests use `cmd/fake-codex` as the backend runner.

```go
import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"
)

type Request struct {
    Args []string
    Env []string
    WorkDir string
    Timeout time.Duration
    Bin string
}

type Response struct {
    ExitCode int
    Stdout string
    Stderr string
    Err error
}

func Setup(t *testing.T, req *Request) error {
    req.WorkDir = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../.."))
    req.Timeout = 30 * time.Second
    if _, err := os.Stat(filepath.Join(req.WorkDir, "go.mod")); err != nil {
        return fmt.Errorf("repo root not resolved: %w", err)
    }

    tmp := t.TempDir()
    doctestBin := filepath.Join(tmp, "doctest")
    build := exec.Command("go", "build", "-o", doctestBin, "./agents/doctest")
    build.Dir = req.WorkDir
    if out, err := build.CombinedOutput(); err != nil {
        t.Fatalf("build doctest: %v\n%s", err, string(out))
    }
    req.Bin = doctestBin
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
    defer cancel()

    bin := req.Bin
    if bin == "" {
        return nil, fmt.Errorf("req.Bin is not set")
    }
    cmd := exec.CommandContext(ctx, bin, req.Args...)
    cmd.Dir = req.WorkDir
    cmd.Env = append(os.Environ(), req.Env...)

    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    resp := &Response{
        Stdout: stdout.String(),
        Stderr: stderr.String(),
        Err: err,
    }
    if err == nil {
        return resp, nil
    }
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        resp.ExitCode = exitErr.ExitCode()
        return resp, nil
    }
    if ctx.Err() != nil {
        return resp, ctx.Err()
    }
    return resp, err
}
```
