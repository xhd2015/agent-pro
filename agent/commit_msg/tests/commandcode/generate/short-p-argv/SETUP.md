# Scenario

**Feature**: commandcode delivers full commit prompt via temp file; `-p` stays short

```
# large staged unified diff must not sit in argv
repo/ (large staged) -> gen-commit-msg --agent-runner commandcode --agent-runner-binary <argv-recorder>
  -> production writes full prompt to temp file (commit-msg-prompt-*.txt)
  -> recorder sees -p <short instruction with path / read …>
  -> -p must NOT embed multi-line "diff --git" / "Git diff:" body
  -> recorder may copy prompt file while agent runs
  -> stdout: title/description from recorder JSON; no --commit
```

## Preconditions
- Isolated git repo with a multi-line staged file (enlarges unified diff).
- Argv recorder binary captures NUL-separated args and best-effort prompt file copy.
- Classic TDD: RED until `runCommandCodeAgent` (shared path) uses temp-file prompt delivery.

## Steps
1. Stage a large text change (`StageCommandCodeRepoLargeDiff`).
2. Install argv recorder as `--agent-runner-binary`.
3. Record HEAD subject; run without `--commit`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	StageCommandCodeRepoLargeDiff(t, req)
	InstallCommandCodeArgvRecorder(t, req)
	req.HEADSubjectBefore = GitHEADSubjectCmd(t, req.GitDir)
	req.Commit = false
	req.DryRun = false
	req.AgentRunner = "commandcode"
	req.Operation = "commandcode-generate-short-p-argv"
	return nil
}
```
