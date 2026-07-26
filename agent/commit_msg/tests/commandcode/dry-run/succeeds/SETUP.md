# Scenario

**Feature**: `--dry-run --agent-runner commandcode` succeeds without calling agent

```
# stage 1 text file; pure plan
1 staged -> gen-commit-msg --dry-run --agent-runner commandcode
  -> stdout: dry-run: would generate commit message for 1 staged file(s)
  -> exit 0; no agent
```

## Preconditions
- One staged text file.
- Non-existent agent binary (must not be invoked).
- Runner ID `commandcode` must be in the supported list (not rejected as unknown).

## Steps
1. Stage one text file in an isolated git repo.
2. Keep DryRun=true and non-existent binary from parent.
3. Run gen-commit-msg without `--commit`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	StageCommandCodeRepo(t, req)
	req.DryRun = true
	req.Commit = false
	req.AgentRunner = "commandcode"
	req.Operation = "commandcode-dry-run-succeeds"
	return nil
}
```
