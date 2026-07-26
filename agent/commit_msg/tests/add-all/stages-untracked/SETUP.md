# Scenario

**Feature**: real `--add-all` stages untracked files before generate (no `--commit`)

```
# only untracked text; empty prior stage
repo/ (untracked.go)
  -> gen-commit-msg --add-all   # no --commit
  -> stderr: $ git add -A
  -> index includes untracked.go
  -> agent mock returns message; stdout has title
  -> HEAD subject unchanged
```

## Preconditions
- Isolated git repo with one untracked text file (not staged).
- fake-opencode mock returns fixed JSON commit message.

## Steps
1. Init repo; write untracked file without staging.
2. Write fake-opencode mock config.
3. Record HEAD subject; set `AddAll` without `Commit`.
4. Run gen-commit-msg.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	name := InitAddAllRepoWithUntracked(t, req)
	req.Operation = name
	req.HEADSubjectBefore = GitHEADSubjectAddAll(t, req.GitDir)

	WriteMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_add_all_stages","llm_events":[{"type":"step_start"},{"type":"message","text":"{\"title\": \"feat: add untracked\", \"description\": \"via --add-all\"}"},{"type":"step_finish"}]}`)

	req.AddAll = true
	req.DryRun = false
	req.Commit = false
	req.AgentRunner = "opencode"
	req.AgentRunnerBinary = req.FakeOpencode
	return nil
}
```
