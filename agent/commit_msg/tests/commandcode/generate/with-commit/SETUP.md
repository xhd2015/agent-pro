# Scenario

**Feature**: generate + `--commit` with commandcode mock binary

```
# stage 1 file; agent path + commit
repo/ (1 staged) -> gen-commit-msg --agent-runner commandcode --agent-runner-binary <mock> --commit
  -> exit 0
  -> new commit subject: feat: via commandcode
```

## Preconditions
- Isolated git repo with hooks disabled (`core.hooksPath=/dev/null` via InitGitRepo).
- One staged text file.
- Mock hook returns fixed JSON title/description.

## Steps
1. Stage a change.
2. Run gen-commit-msg with `--commit`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	StageCommandCodeRepo(t, req)
	req.HEADSubjectBefore = GitHEADSubjectCmd(t, req.GitDir)
	req.Commit = true
	req.DryRun = false
	req.NoVerify = false
	req.AgentRunner = "commandcode"
	req.AgentRunnerBinary = req.MockCommandCode
	req.CommandCodeHook = DefaultCommandCodeHook
	req.Operation = "commandcode-generate-with-commit"
	return nil
}
```
