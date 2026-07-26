# Scenario

**Feature**: pure-plan `--dry-run` accepts `--agent-runner=commandcode`

```
# validation accepts commandcode; pure plan still skips agent
staged -> gen-commit-msg --dry-run --agent-runner commandcode
  -> mock B success; no llm-mock-run-commandcode invoke
```

## Preconditions
- Parent defaults runner to commandcode and builds the mock binary.
- Dry-run success must not require a real agent; leaves may point
  `--agent-runner-binary` at a non-existent path to prove no invoke.

## Steps
1. Inherit commandcode runner ID.
2. Leaves stage a repo, set DryRun, and optionally clear binary path.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DryRun = true
	req.Commit = false
	// Prove agent is not invoked under dry-run even for commandcode.
	req.AgentRunnerBinary = filepath.Join(req.TempDir, "must-not-invoke-commandcode")
	req.CommandCodeHook = ""
	return nil
}
```
