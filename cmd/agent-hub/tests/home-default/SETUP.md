## Preconditions
- The repository contains `cmd/agent-hub`.
- `AGENT_HUB_HOME` is NOT set — test covers the default home directory logic.
- `HOME` is set to a temp directory for isolation.

## Steps
1. Build `cmd/agent-hub`.
2. Set `HOME` to a temp directory and leave `AGENT_HUB_HOME` unset.
3. Set `PATH` so the built binary is reachable.

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
    Home        string
    Command     string
    Args        []string
    Env         []string
    TempDir     string
    AgentHub    string
    RepoRoot    string
    UserHomeDir string
}

type Response struct {
    Stdout   string
    Stderr   string
    ExitCode int
    Err      error
}

func Run(t *testing.T, req *Request) (*Response, error) {
    if req.Command == "" {
        req.Command = req.AgentHub
    }
    return execCmd(t, req.Command, req.Args, req.TempDir, req.Env, "")
}

func Setup(t *testing.T, req *Request) error {
    req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
    if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
        return fmt.Errorf("repo root not found: %w", err)
    }
    req.TempDir = t.TempDir()
    req.UserHomeDir = filepath.Join(req.TempDir, "home")
    if err := os.MkdirAll(req.UserHomeDir, 0755); err != nil {
        return fmt.Errorf("mkdir user home: %w", err)
    }
    req.AgentHub = filepath.Join(req.TempDir, "bin", "agent-hub")

    if err := os.MkdirAll(filepath.Dir(req.AgentHub), 0755); err != nil {
        return fmt.Errorf("mkdir bin: %w", err)
    }

    build := exec.Command("go", "build", "-o", req.AgentHub, "./cmd/agent-hub")
    build.Dir = req.RepoRoot
    if out, err := build.CombinedOutput(); err != nil {
        return fmt.Errorf("build agent-hub: %w\n%s", err, string(out))
    }

    req.Env = append(req.Env,
        "HOME="+req.UserHomeDir,
        "PATH="+filepath.Dir(req.AgentHub)+":"+os.Getenv("PATH"),
    )

    return nil
}

func execCmd(t *testing.T, command string, args []string, dir string, env []string, stdin string) (*Response, error) {
    t.Helper()
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, command, args...)
    cmd.Dir = dir
    cmd.Env = append(os.Environ(), env...)
    if stdin != "" {
        cmd.Stdin = strings.NewReader(stdin)
    }
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()
    resp := &Response{
        Stdout: stdout.String(),
        Stderr: stderr.String(),
        Err:    err,
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
