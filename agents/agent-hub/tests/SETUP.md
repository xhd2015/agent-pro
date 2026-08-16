## Preconditions
- The repository contains `cmd/agent-hub` and `cmd/fake-opencode`.
- Each test runs with an isolated `AGENT_HUB_HOME` directory.
- `AGENT_HUB_OPENCODE_RUNNER` environment variable controls runner redirection.

## Steps
1. Build `cmd/agent-hub` and `cmd/fake-opencode`.
2. Run `agent-hub` or `fake-opencode` with leaf-specific arguments.
3. Capture stdout, stderr, exit code.

```go
import (
	"runtime"
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
    "testing"
    "time"
	"github.com/xhd2015/doctest/session"
)

func execCmd(t *testing.T, command string, args []string, dir string, env []string, stdin string) (*Response, error) {
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

func runAgentHub(t *testing.T, req *Request, args ...string) (*Response, error) {
    t.Helper()
    return execCmd(t, req.AgentHub, args, req.TempDir, req.Env, "")
}

func runFakeOpencode(t *testing.T, req *Request, mockConfigPath string) (*Response, error) {
    t.Helper()
    return execCmd(t, req.FakeOpencode, []string{"run", "--format", "json", "--mock-config", mockConfigPath, "hello"}, req.TempDir, req.Env, "")
}

func writeFile(t *testing.T, path string, content string) {
    t.Helper()
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        t.Fatalf("write %s: %v", path, err)
    }
}

func assertSuccess(t *testing.T, resp *Response) {
    t.Helper()
    if resp.Err != nil && resp.ExitCode == 0 {
        t.Fatalf("run failed: %v", resp.Err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
    }
}

func assertExitCode(t *testing.T, resp *Response, want int) {
    t.Helper()
    if resp.ExitCode != want {
        t.Fatalf("expected exit code %d, got %d, stderr:\n%s", want, resp.ExitCode, resp.Stderr)
    }
}

func assertContains(t *testing.T, got string, want string) {
    t.Helper()
    if !strings.Contains(got, want) {
        t.Fatalf("missing %q in:\n%s", want, got)
    }
}

func parseJSON(t *testing.T, text string) map[string]any {
    t.Helper()
    var obj map[string]any
    if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &obj); err != nil {
        t.Fatalf("invalid JSON: %v\n%s", err, text)
    }
    return obj
}

func parseJSONLines(t *testing.T, text string) []map[string]any {
    t.Helper()
    var out []map[string]any
    for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
        if strings.TrimSpace(line) == "" {
            continue
        }
        var obj map[string]any
        if err := json.Unmarshal([]byte(line), &obj); err != nil {
            t.Fatalf("invalid JSON line: %v\n%s", err, line)
        }
        out = append(out, obj)
    }
    return out
}

func notifyEvent(t *testing.T, req *Request, eventType, runner, sessionID string) map[string]any {
    t.Helper()
    body := fmt.Sprintf(`{"event_type":"%s","runner":"%s","runner_session_id":"%s"}`, eventType, runner, sessionID)
    prevEnvs := req.Env
    req.Env = append([]string{}, req.Env...)
    resp, err := runAgentHub(t, req, "notify", "--json", body)
    req.Env = prevEnvs
    if err != nil {
        t.Fatalf("notifyEvent error: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("notifyEvent failed (exit %d):\nstderr: %s", resp.ExitCode, resp.Stderr)
    }
    return parseJSON(t, resp.Stdout)
}

func toInt(v any) (int64, bool) {
    switch n := v.(type) {
    case float64:
        return int64(n), true
    case int64:
        return n, true
    case string:
        i, err := strconv.ParseInt(n, 10, 64)
        if err != nil {
            return 0, false
        }
        return i, true
    }
    return 0, false
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../.."))
    if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
        return fmt.Errorf("repo root not found: %w", err)
    }
    req.TempDir = t.TempDir()
    req.Home = filepath.Join(req.TempDir, "agent-hub-home")
    req.AgentHub = filepath.Join(req.TempDir, "bin", "agent-hub")
    req.FakeOpencode = filepath.Join(req.TempDir, "bin", "fake-opencode")

    if err := os.MkdirAll(filepath.Dir(req.AgentHub), 0755); err != nil {
        return fmt.Errorf("mkdir bin: %w", err)
    }

    build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.AgentHub, "./cmd/agent-hub")
    build.Dir = req.RepoRoot
    if out, err := build.CombinedOutput(); err != nil {
        return fmt.Errorf("build agent-hub: %w\n%s", err, string(out))
    }

    build2 := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.FakeOpencode, "./cmd/fake-opencode")
    build2.Dir = req.RepoRoot
    if out, err := build2.CombinedOutput(); err != nil {
        return fmt.Errorf("build fake-opencode: %w\n%s", err, string(out))
    }

    req.Env = append(req.Env,
        "AGENT_HUB_HOME="+req.Home,
        "PATH="+filepath.Dir(req.AgentHub)+":"+os.Getenv("PATH"),
    )
    return nil
}

func runFullWorkflow(t *testing.T, req *Request) (*Response, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, req.Command, req.Args...)
    cmd.Dir = req.TempDir
    cmd.Env = append(os.Environ(), req.Env...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    mkErrResp := func(format string, args ...any) (*Response, error) {
        msg := fmt.Sprintf(format, args...)
        return &Response{Stdout: stdout.String(), Stderr: stderr.String() + msg, Err: fmt.Errorf("%s", msg)}, fmt.Errorf("%s", msg)
    }

    if err := cmd.Start(); err != nil {
        return mkErrResp("cmd.Start: %v", err)
    }

    time.Sleep(1000 * time.Millisecond)

    sr, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "sess_full")
    if err != nil {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("session show (mid-flight): %v", err)
    }
    if sr.ExitCode != 0 {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("session show failed (exit %d): %s", sr.ExitCode, sr.Stderr)
    }
    so := parseJSON(t, sr.Stdout)
    if so["status"] != "running" {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("expected session status running mid-flight, got %v", so["status"])
    }

    fr, err := runAgentHub(t, req, "fetch", "--consumer-id", "consumer-full-"+t.Name(), "--limit", "10")
    if err != nil {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("fetch (mid-flight): %v", err)
    }
    if fr.ExitCode != 0 {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("fetch failed (exit %d): %s", fr.ExitCode, fr.Stderr)
    }
    frObj := parseJSON(t, fr.Stdout)
    events, _ := frObj["events"].([]any)
    if events == nil || len(events) == 0 {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("expected at least 1 event mid-flight, got 0")
    }
    hasStarted := false
    for _, e := range events {
        ev := e.(map[string]any)
        nev, _ := ev["event"].(map[string]any)
        if nev["event_type"] == "agent.session.started" {
            hasStarted = true
        }
    }
    if !hasStarted {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("expected agent.session.started event mid-flight")
    }

    err = cmd.Wait()
    resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
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
