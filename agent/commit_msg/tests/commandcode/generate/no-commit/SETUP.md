# Scenario

**Feature**: generate message with commandcode mock; print only (no `--commit`)

```
# stage 1 file; agent path; no commit
repo/ (1 staged) -> gen-commit-msg --agent-runner commandcode --agent-runner-binary <mock>
  -> exit 0
  -> stdout contains "feat: via commandcode" and "from commandcode mock"
  -> HEAD subject unchanged
```

## Preconditions
- Isolated git repo with one staged text file.
- Mock binary + `LLM_MOCK_RUN_COMMANDCODE_COMMAND` return fixed JSON.

## Steps
1. Stage `feature.go`.
2. Record HEAD subject before run.
3. Run without `--commit`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	StageCommandCodeRepo(t, req)
	req.HEADSubjectBefore = GitHEADSubjectCmd(t, req.GitDir)
	req.Commit = false
	req.DryRun = false
	req.AgentRunner = "commandcode"
	req.AgentRunnerBinary = req.MockCommandCode
	req.CommandCodeHook = DefaultCommandCodeHook
	req.Operation = "commandcode-generate-no-commit"
	return nil
}
```
