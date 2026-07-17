# Scenario

**Feature**: agent-pro exposes debug-with-user via knownSkills registration

```
go build cmd/agent-pro -> agent-pro skill debug-with-user --show -> embedded SKILL.md
```

## Preconditions

- `cmd/agent-pro/skill_cmd.go` registers `debug-with-user` in `knownSkills`.
- `knownSkillNames()` includes `debug-with-user`.

## Steps

1. Resolve repo root from `DOCTEST_ROOT/../../../..`.
2. Build `agent-pro` binary into temp dir.
3. Leaf runs `skill debug-with-user show`.

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
	build := exec.Command("go", "build", "-buildvcs=false", "-o", req.AgentPro, "./cmd/agent-pro")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-pro: %w\n%s", err, string(out))
	}
	return nil
}
```
