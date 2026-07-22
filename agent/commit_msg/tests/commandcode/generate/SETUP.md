# Scenario

**Feature**: generate commit message with commandcode mock binary

```
# agent path (not dry-run)
staged -> gen-commit-msg --agent-runner commandcode --agent-runner-binary <llm-mock|argv-recorder>
  -> short -p + temp prompt file (shared delivery) OR legacy full -p until implementer
  -> mock/recorder prints fixed JSON
  -> stdout: title + description; optional --commit
```

## Preconditions
- Parent built `llm-mock-run-commandcode` and set default CommandCodeHook.
- Leaves stage a repo and set Commit true/false, or install an argv recorder.

## Steps
1. Inherit runner + mock binary + hook from parent.
2. Leaves stage git changes and set Commit flag (or replace binary with argv recorder).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DryRun = false
	req.AgentRunner = "commandcode"
	// Leaves that install an argv recorder set AgentRunnerBinary themselves and
	// clear CommandCodeHook; do not overwrite a non-mock binary here.
	if req.AgentRunnerBinary == "" || req.AgentRunnerBinary == req.MockCommandCode {
		if req.MockCommandCode != "" {
			req.AgentRunnerBinary = req.MockCommandCode
		}
		if req.CommandCodeHook == "" {
			req.CommandCodeHook = DefaultCommandCodeHook
		}
	}
	return nil
}
```
