# Scenario

**Feature**: `--add-all --dry-run` plans `git add -A` without mutating the index

```
# empty index + untracked file under dry-run + add-all
repo/ (untracked.go untracked, empty index)
  -> gen-commit-msg --add-all --dry-run
  -> stderr: would: git add -A
  -> index still empty (untracked stays unstaged)
  -> error: no staged changes … (honest dry-run; count from current index)
```

## Preconditions
- Isolated git repo with initial commit only.
- One untracked text file; nothing newly staged.
- Agent binary path is non-existent (must not run under pure dry-run).

## Steps
1. Init repo and write untracked file without `git add`.
2. Set `AddAll` + `DryRun`.
3. Run gen-commit-msg.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	name := InitAddAllRepoWithUntracked(t, req)
	req.Operation = name
	req.AddAll = true
	req.DryRun = true
	req.Commit = false
	// Non-existent agent proves dry-run must not invoke the runner.
	req.AgentRunnerBinary = filepath.Join(req.TempDir, "must-not-invoke-agent-add-all")
	return nil
}
```
