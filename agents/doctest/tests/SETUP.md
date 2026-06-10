## Preconditions
- The repository root contains the doctest command at `./agents/doctest`.
- The tests are executed by the doc-style test runner from this test tree.
- Each leaf sets the doctest arguments it wants to execute.

## Steps
1. Resolve the repository root from `DOCTEST_ROOT`.
2. Execute `DOCTEST_BIN` when provided, otherwise execute `go run ./agents/doctest`.
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
    "strings"
    "testing"
    "time"
)

type Request struct {
    Args []string
    Env []string
    WorkDir string
    Timeout time.Duration
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
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
    defer cancel()

    repoRoot := filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../.."))
    bin := strings.TrimSpace(os.Getenv("DOCTEST_BIN"))
    var cmd *exec.Cmd
    if bin != "" {
        cmd = exec.CommandContext(ctx, bin, req.Args...)
    } else {
        args := append([]string{"run", filepath.Join(repoRoot, "agents/doctest")}, req.Args...)
        cmd = exec.CommandContext(ctx, "go", args...)
    }
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
