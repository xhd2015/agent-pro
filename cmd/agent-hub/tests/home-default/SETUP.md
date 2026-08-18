# Scenario

**Feature**: agent-hub home resolution uses isolated HOME / AGENT_HUB_HOME on the child process

```
# harness builds agent-hub, sets child Env (HOME, PATH)
doctest -> go build -C cmd ./agent-hub -> child agent-hub daemon status

# default home or AGENT_HUB_HOME override (no parent os.Setenv)
doctest <- JSON home + running=false
```

## Preconditions

- The repository contains `cmd/agent-hub`.
- Default-home leaves leave `AGENT_HUB_HOME` unset on the child.
- `HOME` is set on the **child** env to a temp directory for isolation (not process Setenv).

## Steps

1. Resolve module root from `d.DOCTEST_ROOT`.
2. Build `cmd/agent-hub` into a leaf temp `bin/`.
3. Seed `req.Env` with isolated `HOME` and PATH prefix so the binary is found.

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
	// Tree root: cmd/agent-hub/tests/home-default → module root is four levels up.
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
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

	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", req.AgentHub, "./agent-hub")
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
```
