## Preconditions
- A message event carries `delay_ms: 2000`.

## Steps
1. Write mock config with a message event that has a 2s pre-emission delay.
2. Run fake-opencode. The overridden Run measures elapsed wall time.

```go
import (
    "bytes"
    "context"
    "errors"
    "testing"
    "time"
    "os/exec"
    "os"
)

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_delay","llm_events":[{"type":"message","text":"delayed","delay_ms":2000}]}`)
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    start := time.Now()
    args := []string{"run", "--format", "json", "--mock-config", req.MockConfigPath, "hello"}
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, req.FakeOpencode, args...)
    cmd.Dir = req.TempDir
    cmd.Env = append(os.Environ(), req.Env...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()
    elapsed := time.Since(start)
    resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
    if err == nil {
        if elapsed < 1500*time.Millisecond {
            t.Fatalf("expected at least 1500ms elapsed due to delay_ms=2000, got %v", elapsed)
        }
        return resp, nil
    }
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        resp.ExitCode = exitErr.ExitCode()
        if elapsed < 1500*time.Millisecond {
            t.Fatalf("expected at least 1500ms elapsed due to delay_ms=2000, got %v", elapsed)
        }
        return resp, nil
    }
    if ctx.Err() != nil {
        return resp, ctx.Err()
    }
    return resp, err
}
```
