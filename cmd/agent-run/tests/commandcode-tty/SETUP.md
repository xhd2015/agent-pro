# Scenario

Builds `agent-run` and `llm-mock-run-commandcode` into temp dir. Sets `AGENT_RUN_HOME`.

## Steps

1. Resolve repo root from `DOCTEST_ROOT/../../../..`.
2. Build `agent-run` binary.
3. Build `llm-mock-run-commandcode` binary (unless `SkipMockBuild`).
4. Set `AGENT_RUN_HOME` to temp directory.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}

	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-pro")
	t.Setenv("AGENT_RUN_HOME", req.Home)

	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	build := exec.Command("go", "build", "-buildvcs=false", "-o", req.AgentRun, "./cmd/agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build agent-run: %v\n%s", err, out)
	}

	if !req.SkipMockBuild {
		req.MockBinary = filepath.Join(req.TempDir, "bin", "llm-mock-run-commandcode")
		build2 := exec.Command("go", "build", "-buildvcs=false", "-o", req.MockBinary,
			"./agent/llm/llm-mock/llm-mock-run-commandcode")
		build2.Dir = req.RepoRoot
		if out2, err2 := build2.CombinedOutput(); err2 != nil {
			t.Fatalf("build mock: %v\n%s", err2, out2)
		}
	}

	return nil
}
```
