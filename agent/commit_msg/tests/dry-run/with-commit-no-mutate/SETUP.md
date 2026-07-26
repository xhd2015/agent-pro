# Scenario

**Feature**: `--dry-run --commit` plans git commit without mutating HEAD

```
# dry-run + commit flag: would-line only
staged -> gen-commit-msg --dry-run --commit
  -> stderr: would: git commit -m '…'
  -> HEAD subject unchanged
```

## Preconditions
- One staged text change; HEAD subject is the initial commit subject.
- Agent binary is non-existent (must not run).

## Steps
1. Stage one text file in an isolated repo.
2. Record that dry-run + commit is requested.
3. Run gen-commit-msg with `--dry-run --commit`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	StageDryRunRepo(t, req, 1)
	req.DryRun = true
	req.Commit = true
	req.NoVerify = false
	req.AgentRunnerBinary = NonExistentAgentBinary(req)
	req.Operation = GitHEADSubject(t, req.GitDir)
	return nil
}
```
