## Preconditions
- The repository contains or will contain `cmd/fake-opencode`.
- Each test runs with a temporary HOME and opencode config directory.

## Steps
1. Build `cmd/fake-opencode`.
2. Write a leaf-specific mock config when needed.
3. Run fake opencode with leaf-specific arguments.
4. Capture stdout, stderr, exit status, hook log, and host config side effects.

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

func Setup(t *testing.T, req *Request) error {
    _ = writeHookRecorder
    _ = writeMockConfig
    _ = assertSuccess
    _ = assertContains
    _ = parseJSONLines
    req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../.."))
    if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
        return fmt.Errorf("repo root not found: %w", err)
    }
    req.TempDir = t.TempDir()
    req.HomeDir = filepath.Join(req.TempDir, "home")
    req.FakeOpencode = filepath.Join(req.TempDir, "fake-opencode")
    req.HookLogPath = filepath.Join(req.TempDir, "hooks.jsonl")
    req.MarkerPath = filepath.Join(req.TempDir, "markers.log")
    build := exec.Command("go", "build", "-buildvcs=false", "-o", req.FakeOpencode, "./cmd/fake-opencode")
    build.Dir = req.RepoRoot
    if out, err := build.CombinedOutput(); err != nil {
        return fmt.Errorf("build fake-opencode: %w\n%s", err, string(out))
    }
    req.Env = append(req.Env,
        "HOME="+req.HomeDir,
        "OPENCODE_CONFIG_DIR="+filepath.Join(req.TempDir, "opencode-config"),
    )
    return nil
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

func writeMockConfig(t *testing.T, req *Request, body string) {
    t.Helper()
    req.MockConfigPath = filepath.Join(req.TempDir, "mock.json")
    writeFile(t, req.MockConfigPath, body)
}

func writeHookRecorder(t *testing.T, req *Request, exitCode int) string {
    t.Helper()
    script := filepath.Join(req.TempDir, "hook-recorder.sh")
    content := fmt.Sprintf(`#!/bin/sh
event="$1"
payload="$(cat)"
printf '{"event":"%%s","payload":%%s}\n' "$event" "$payload" >> %q
exit %d
`, req.HookLogPath, exitCode)
    writeFile(t, script, content)
    if err := os.Chmod(script, 0755); err != nil {
        t.Fatalf("chmod hook recorder: %v", err)
    }
    return script
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

func assertContains(t *testing.T, got string, want string) {
    t.Helper()
    if !strings.Contains(got, want) {
        t.Fatalf("missing %q in:\n%s", want, got)
    }
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

func runPerEventDelay(t *testing.T, req *Request) (*Response, error) {
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
    if ctx.Err() != nil {
        return resp, ctx.Err()
    }
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        resp.ExitCode = exitErr.ExitCode()
        if elapsed < 1500*time.Millisecond {
            t.Fatalf("expected at least 1500ms elapsed due to delay_ms=2000, got %v", elapsed)
        }
        return resp, nil
    }
    return resp, err
}
```
