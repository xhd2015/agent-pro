# Scenario

**Feature**: unknown `--agent-runner` is rejected even with `--dry-run`

```
# validation before pure-plan success
gen-commit-msg --dry-run --agent-runner codex
  -> error: unsupported agent runner: codex (supported: opencode)
```

## Preconditions
- Staged changes optional; runner validation fails before mock success.
- Use runner name `codex` (not supported by gen-commit-msg).

## Steps
1. Stage one file so a successful dry-run would otherwise proceed.
2. Set `req.AgentRunner = "codex"` and `req.DryRun = true`.
3. Run gen-commit-msg.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	StageDryRunRepo(t, req, 1)
	req.DryRun = true
	req.Commit = false
	req.AgentRunner = "codex"
	req.AgentRunnerBinary = NonExistentAgentBinary(req)
	return nil
}
```
