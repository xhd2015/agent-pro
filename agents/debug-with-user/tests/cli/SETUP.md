# Scenario

**Feature**: debug-with-user CLI subprocess tests with dry-run (no GUI)

```
go build agents/debug-with-user -> temp binary -> exec ask/skill show -> capture stdout/stderr/exit
```

## Preconditions

- `agents/debug-with-user/main.go` implements `ask` and `skill show`.
- Dry-run env vars avoid `osascript` in CI.

## Steps

1. Resolve repo root from `d.DOCTEST_ROOT/../../../..`.
2. Build `debug-with-user` into `t.TempDir()/bin/`.
3. Leaf setup sets `req.Args` and dry-run `req.Env`.
4. `Run` executes the binary and captures output.

## Context

- `d.DOCTEST_ROOT` for this nested root is `agents/debug-with-user/tests/cli`.
- Module root is `d.DOCTEST_ROOT/../../../..` (four levels up from nested cli root).

```go
import (
	"runtime"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found at %s: %w", req.RepoRoot, err)
	}
	req.TempDir = t.TempDir()
	req.Binary = filepath.Join(req.TempDir, "bin", "debug-with-user")
	if err := os.MkdirAll(filepath.Dir(req.Binary), 0o755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.Binary, "./agents/debug-with-user")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build debug-with-user: %w\n%s", err, string(out))
	}
	return nil
}
```
