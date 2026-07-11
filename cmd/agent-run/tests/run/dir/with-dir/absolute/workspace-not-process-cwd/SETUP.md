# Scenario

**Feature**: `run --dir <absDir>` from another process cwd records absDir as workspace

```
# process cwd = TempDir (other); workspace = TempDir/ws-target
agent-run run --dir <abs(ws-target)> --json --agent-runner fake-codex "hi"
  -> exit 0
  -> meta.workspace = abs(ws-target) (or symlink-canonical form)
  -> meta.workspace ≠ process cwd
```

## Steps

1. Create `ws-target` under TempDir.
2. Run with absolute `--dir` and a short fake-codex prompt.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	ws := filepath.Join(req.TempDir, "ws-target")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	abs, err := filepath.Abs(ws)
	if err != nil {
		return err
	}
	// process cwd remains req.TempDir (root runAgentRun cmd.Dir)
	req.Args = append(req.Args, "--dir", abs, "--json", "hi")
	return nil
}
```
