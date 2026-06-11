## Preconditions
- The repository contains `agents/doctest` and `cmd/fake-codex`.
- Each test runs in a temporary directory.
- The doctest binary is built from source.
- The fake-codex binary is built from source for sub-agent simulation.

## Steps
1. Build `doctest` binary into the test temporary directory.
2. Build `fake-codex` binary into the test temporary directory.
3. Copy doctest as `yield-pending-questions` in the same temp dir.
4. Run with leaf-specific args and env.
5. Capture stdout, stderr, and exit status.

## Context
- The root harness provides helpers for writing mock configs and assertions.
- Sub-agent behavior is simulated via `fake-codex --json --mock-config`.
- The default Run wraps `doctest agent implement` using `--agent-runner fake-codex`.

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
)

type Request struct {
    RepoRoot      string
    TempDir       string
    DoctestBin    string
    FakeCodexBin  string
    YieldPQBin    string
    Args          []string
    Env           []string
    MockConfigPath string

    // If true, run the yield-pending-questions binary directly
    // instead of doctest agent implement.
    AsYieldPQ     bool
}

type Response struct {
    ExitCode int
    Stdout   string
    Stderr   string
    FifoData string // read from FIFO for yield-pending-questions tests
    Err      error
}

func Setup(t *testing.T, req *Request) error {
    _ = writeMockConfig
    _ = assertSuccess
    _ = assertContains
    _ = assertNotContains
    _ = assertExitCode
    _ = writeFile

    req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
    if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
        return fmt.Errorf("repo root not found: %w", err)
    }

    req.TempDir = t.TempDir()
    req.DoctestBin = filepath.Join(req.TempDir, "doctest")
    req.FakeCodexBin = filepath.Join(req.TempDir, "fake-codex")
    req.YieldPQBin = filepath.Join(req.TempDir, "yield-pending-questions")

    // Build doctest
    buildDoctest := exec.Command("go", "build", "-o", req.DoctestBin, "./agents/doctest")
    buildDoctest.Dir = req.RepoRoot
    if out, err := buildDoctest.CombinedOutput(); err != nil {
        return fmt.Errorf("build doctest: %w\n%s", err, string(out))
    }

    // Build fake-codex
    buildFC := exec.Command("go", "build", "-o", req.FakeCodexBin, "./cmd/fake-codex")
    buildFC.Dir = req.RepoRoot
    if out, err := buildFC.CombinedOutput(); err != nil {
        return fmt.Errorf("build fake-codex: %w\n%s", err, string(out))
    }

    // Copy doctest as yield-pending-questions
    if out, err := exec.Command("cp", req.DoctestBin, req.YieldPQBin).CombinedOutput(); err != nil {
        return fmt.Errorf("copy yield-pq: %w\n%s", err, string(out))
    }

    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if req.AsYieldPQ {
        return runYieldPQ(t, ctx, req)
    }
    return runDoctestAgent(t, ctx, req)
}

func runDoctestAgent(t *testing.T, ctx context.Context, req *Request) (*Response, error) {
    args := append([]string{"agent", "implement"}, req.Args...)
    cmd := exec.CommandContext(ctx, req.DoctestBin, args...)
    cmd.Dir = req.TempDir
    cmd.Env = append(os.Environ(), req.Env...)
    cmd.Env = append(cmd.Env,
        "AGENT_RUNNER_FAKE_CODEX_PATH="+req.FakeCodexBin,
        "PATH="+req.TempDir+string(filepath.ListSeparator)+os.Getenv("PATH"),
    )

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()

    resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
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

func runYieldPQ(t *testing.T, ctx context.Context, req *Request) (*Response, error) {
    cmd := exec.CommandContext(ctx, req.YieldPQBin, req.Args...)
    cmd.Dir = req.TempDir
    cmd.Env = append(os.Environ(), req.Env...)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()

    resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
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

func createFifo(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "questions.jsonl")
	return path
}

func readFifo(t *testing.T, path string, timeout time.Duration) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read questions file: %v", err)
	}
	return string(data)
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

func assertSuccess(t *testing.T, resp *Response) {
    t.Helper()
    if resp.Err != nil && resp.ExitCode == 0 {
        t.Fatalf("run failed: %v", resp.Err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d, want 0\nstderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
    }
}

func assertExitCode(t *testing.T, resp *Response, want int) {
    t.Helper()
    if resp.ExitCode != want {
        t.Fatalf("exit code = %d, want %d\nstderr:\n%s\nstdout:\n%s", resp.ExitCode, want, resp.Stderr, resp.Stdout)
    }
}

func assertContains(t *testing.T, got string, want string) {
    t.Helper()
    if !strings.Contains(got, want) {
        t.Fatalf("missing %q in:\n%s", want, got)
    }
}

func assertNotContains(t *testing.T, got string, want string) {
    t.Helper()
    if strings.Contains(got, want) {
        t.Fatalf("unexpected %q in:\n%s", want, got)
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
```
