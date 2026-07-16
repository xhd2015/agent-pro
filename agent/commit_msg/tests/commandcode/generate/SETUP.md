# Scenario

**Feature**: generate commit message with commandcode mock binary

```
# agent path (not dry-run)
staged -> gen-commit-msg --agent-runner commandcode --agent-runner-binary <llm-mock-run-commandcode>
  -> LLM_MOCK_RUN_COMMANDCODE_COMMAND prints fixed JSON
  -> stdout: title + description; optional --commit
```

## Preconditions
- Parent built `llm-mock-run-commandcode` and set default CommandCodeHook.
- Leaves stage a repo and set Commit true/false.

## Steps
1. Inherit runner + mock binary + hook from parent.
2. Leaves stage git changes and set Commit flag.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DryRun = false
	req.AgentRunner = "commandcode"
	if req.MockCommandCode != "" {
		req.AgentRunnerBinary = req.MockCommandCode
	}
	if req.CommandCodeHook == "" {
		req.CommandCodeHook = DefaultCommandCodeHook
	}
	return nil
}
```
