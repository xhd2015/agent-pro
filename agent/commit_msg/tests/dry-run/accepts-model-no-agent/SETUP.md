# Scenario

**Feature**: `--dry-run --model` accepts model without calling the agent

```
# model flag is accepted, unused for generation under dry-run
staged -> gen-commit-msg --dry-run --model some/model
  -> mock B success; no agent / no model-required failure
```

## Preconditions
- One staged text file.
- Explicit non-default model string.
- Non-existent agent binary.

## Steps
1. Stage one text file.
2. Set DryRun and Model to `some/model`.
3. Run gen-commit-msg.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	StageDryRunRepo(t, req, 1)
	req.DryRun = true
	req.Commit = false
	req.Model = "some/model"
	req.AgentRunnerBinary = NonExistentAgentBinary(req)
	return nil
}
```
