# Scenario

**Feature**: agent-pro exposes establish-a-loop via knownSkills registration

```
go build cmd/agent-pro -> agent-pro skill establish-a-loop show -> embedded SKILL.md
agent-pro skills -> lists establish-a-loop with description
```

## Preconditions

- `cmd/agent-pro/skill_cmd.go` registers `establish-a-loop` in `knownSkills`.
- `knownSkillNames()` includes `establish-a-loop`.

## Steps

1. Resolve repo root from `DOCTEST_ROOT/../../../..`.
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
)

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
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