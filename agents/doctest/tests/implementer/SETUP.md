## Preconditions
- The `doctest` and `fake-codex` binaries are built into temp dir.
- The `yield-pending-questions` dispatch binary is available.
- The doctest agent implement command uses fake-codex via env vars.

## Steps
1. Build `cmd/fake-codex` into a temporary executable.
2. Build `agents/doctest` into a temporary executable.
3. Copy doctest as `yield-pending-questions`.
4. Configure env vars so agent implement uses fake-codex.

```go
import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"
)

func Setup(t *testing.T, req *Request) error {
    tmp := t.TempDir()

    fakeCodex := filepath.Join(tmp, "fake-codex")
    build := exec.Command("go", "build", "-o", fakeCodex, "./cmd/fake-codex")
    build.Dir = req.WorkDir
    if out, err := build.CombinedOutput(); err != nil {
        t.Fatalf("build fake-codex: %v\n%s", err, string(out))
    }

    doctestBin := filepath.Join(tmp, "doctest")
    buildDT := exec.Command("go", "build", "-o", doctestBin, "./agents/doctest")
    buildDT.Dir = req.WorkDir
    if out, err := buildDT.CombinedOutput(); err != nil {
        t.Fatalf("build doctest: %v\n%s", err, string(out))
    }

    yieldPQ := filepath.Join(tmp, "yield-pending-questions")
    if out, err := exec.Command("cp", doctestBin, yieldPQ).CombinedOutput(); err != nil {
        t.Fatalf("copy yield-pending-questions: %v\n%s", err, string(out))
    }

    req.Env = append(req.Env,
        "DOCTEST_BIN="+doctestBin,
        "AGENT_RUNNER_FAKE_CODEX_PATH="+fakeCodex,
        "YIELD_PQ_BIN="+yieldPQ,
    )
    os.Setenv("YIELD_PQ_BIN", yieldPQ)
    req.Timeout = 60 * time.Second
    return nil
}

func writeMockConfig(t *testing.T, req *Request, body string) string {
    t.Helper()
    path := filepath.Join(t.TempDir(), "mock.json")
    if err := os.WriteFile(path, []byte(body), 0644); err != nil {
        t.Fatalf("write mock config: %v", err)
    }
    req.Env = append(req.Env, "FAKE_CODEX_MOCK_CONFIG="+path)
    return path
}
```
