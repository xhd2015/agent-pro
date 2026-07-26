# Scenario

**Feature**: gen-commit-msg fails early when `--dir` is not inside a git repository

```
# plain directory without .git
non-git --dir -> gen-commit-msg [--dry-run]
  -> error: not a git repository: <dir>
```

## Preconditions
- Target path exists but has no `.git` (not a git work tree).
- Early `IsInsideGit` check runs before staged-diff inspection and before the agent.

## Steps
1. Create an empty non-git directory under the leaf temp dir.
2. Point `req.GitDir` at that path (`--dir`).
3. Optionally set `DryRun` so the pure-plan path is attempted; fail still occurs at git detection.
4. Run gen-commit-msg via the shared capture helper.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	dir := filepath.Join(req.TempDir, "plain")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	req.GitDir = dir
	// Optional: dry-run still hits IsInsideGit before staged/agent work.
	req.DryRun = true
	req.Commit = false
	return nil
}
```
