# Proof: HOME Env Overrides Home Directory

These tests demonstrate that setting the `HOME` environment variable overrides
the home directory across multiple runtimes:

- **bash** — `$HOME` variable, `echo $HOME`
- **golang** — `os.UserHomeDir()`
- **nodejs** — `require('os').homedir()`
- **bun** — `require('os').homedir()`

Each test sets `HOME` to a temporary directory and verifies the runtime
resolves that directory as the user's home — confirming that `HOME` is the
authoritative source for home directory discovery.

These tests do **not** test any code in this project; they are purely
demonstrations of platform behavior.

## How to Run

```sh
doctest test ./tests/proof-of-home-dir
```

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
	"github.com/xhd2015/doctest/session"
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

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
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
```
