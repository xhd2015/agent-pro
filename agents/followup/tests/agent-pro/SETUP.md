# Scenario

**Feature**: agent-pro exposes followup via knownSkills registration

```
go build cmd/agent-pro -> agent-pro skill followup --show -> embedded SKILL.md
agent-pro skills -> lists followup with description
```

## Preconditions

- `cmd/agent-pro/skill_cmd.go` registers `followup` in `knownSkills`.
- `knownSkillNames()` includes `followup`.

## Steps

1. Resolve repo root from `d.DOCTEST_ROOT/../../../..`.
2. Create stub `frontend/dist/index.html` before build.
3. Build `agent-pro` binary into temp dir.
4. Leaf runs the scenario-specific `agent-pro` command.

```go
import (
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
	req.AgentPro = filepath.Join(req.TempDir, "bin", "agent-pro")
	if err := os.MkdirAll(filepath.Dir(req.AgentPro), 0o755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	distDir := filepath.Join(req.RepoRoot, "frontend", "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		return fmt.Errorf("mkdir frontend/dist: %w", err)
	}
	stub := filepath.Join(distDir, "index.html")
	if _, err := os.Stat(stub); os.IsNotExist(err) {
		if err := os.WriteFile(stub, []byte("<!doctype html><title>stub</title>"), 0o644); err != nil {
			return fmt.Errorf("write frontend/dist stub: %w", err)
		}
	}
	build := exec.Command("go", "build", "-o", req.AgentPro, "./cmd/agent-pro")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-pro: %w\n%s", err, string(out))
	}
	return nil
}
```