# Scenario

**Feature**: `--dry-run` prints mock message B with correct staged file count

```
# stage 2 text files, pure plan, no agent
2 staged files -> gen-commit-msg --dry-run
  -> stdout: dry-run: would generate commit message for 2 staged file(s)
```

## Preconditions
- Two known text files are staged in an isolated git repo.
- Agent binary path is intentionally non-existent (must not be invoked).

## Steps
1. Initialize a git repo and stage `change_1.go` and `change_2.go` (N=2).
2. Set `req.DryRun = true` and a non-existent `--agent-runner-binary`.
3. Run gen-commit-msg without `--commit`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	StageDryRunRepo(t, req, 2)
	req.DryRun = true
	req.Commit = false
	req.AgentRunnerBinary = NonExistentAgentBinary(req)
	req.Operation = "mock-message-count"
	return nil
}
```
