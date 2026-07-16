# Scenario

**Feature**: `--add-all --commit` stages untracked then creates a commit

```
# untracked only → add-all + commit with fake-opencode
repo/ (untracked.go)
  -> gen-commit-msg --add-all --commit
  -> stderr: $ git add -A
  -> new HEAD subject from mock title
  -> untracked.go is in the new commit
```

## Preconditions
- Isolated git repo with one untracked text file (not staged).
- fake-opencode mock returns fixed JSON title/description.
- Hooks disabled via InitGitRepo (`core.hooksPath=/dev/null`).

## Steps
1. Init repo; write untracked file without staging.
2. Write fake-opencode mock config.
3. Record HEAD subject; set `AddAll` + `Commit`.
4. Run gen-commit-msg.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	name := InitAddAllRepoWithUntracked(t, req)
	req.Operation = name
	req.HEADSubjectBefore = GitHEADSubjectAddAll(t, req.GitDir)

	WriteMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_add_all_commit","llm_events":[{"type":"step_start"},{"type":"message","text":"{\"title\": \"feat: commit untracked\", \"description\": \"via --add-all --commit\"}"},{"type":"step_finish"}]}`)

	req.AddAll = true
	req.DryRun = false
	req.Commit = true
	req.NoVerify = false
	req.AgentRunner = "opencode"
	req.AgentRunnerBinary = req.FakeOpencode
	return nil
}
```
