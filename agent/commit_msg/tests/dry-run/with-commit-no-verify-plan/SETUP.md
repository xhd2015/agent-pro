# Scenario

**Feature**: `--dry-run --commit --no-verify` plans commit with `--no-verify` on stderr

```
# dry-run commit plan includes --no-verify
staged -> gen-commit-msg --dry-run --commit --no-verify
  -> stderr would-line includes --no-verify
  -> HEAD unchanged
```

## Preconditions
- One staged text change.
- Agent binary non-existent.

## Steps
1. Stage one text file.
2. Set DryRun, Commit, and NoVerify.
3. Run gen-commit-msg.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	StageDryRunRepo(t, req, 1)
	req.DryRun = true
	req.Commit = true
	req.NoVerify = true
	req.AgentRunnerBinary = NonExistentAgentBinary(req)
	req.Operation = GitHEADSubject(t, req.GitDir)
	return nil
}
```
