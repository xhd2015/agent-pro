# Scenario

**Feature**: `--dry-run` with empty index errors like the normal path

```
# empty staged set under dry-run
clean index -> gen-commit-msg --dry-run
  -> error: no staged changes …
```

## Preconditions
- Git repo exists with an initial commit but nothing newly staged.
- Agent binary non-existent.

## Steps
1. Initialize a git repo only (no additional staged files).
2. Set DryRun.
3. Run gen-commit-msg.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.GitDir = filepath.Join(req.TempDir, "repo")
	InitGitRepo(t, req.GitDir)
	req.DryRun = true
	req.Commit = false
	req.AgentRunnerBinary = NonExistentAgentBinary(req)
	return nil
}
```
