## Preconditions
- A valid doc-style test tree exists in the repository.
- The command is invoked from inside a nested repository directory.

## Steps
1. Set the process working directory to `agents/test-case-tree-runner`.
2. Run `doctest test <absolute-dir>`.

```go
import (
    "bytes"
    "context"
    "errors"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    repoRoot := req.WorkDir
    exampleDir := filepath.Join(repoRoot, "agents/test-case-tree-runner/examples/basic-request-runner")
    req.WorkDir = filepath.Join(repoRoot, "agents/test-case-tree-runner")
    req.Args = []string{"test", exampleDir}
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
