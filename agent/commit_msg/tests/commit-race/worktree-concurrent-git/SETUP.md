# Scenario

Same index.lock race as background-git-loop, but from a linked git worktree.

## Preconditions
- A main git repo with a linked worktree (mirrors the reported `agent-pro-task-hub` failure).
- fake-opencode spawns a background git loop in the worktree before returning the commit message.
- gen-commit-msg runs with `--commit` from the worktree directory.

## Steps
1. Initialize main repo and add a worktree.
2. Stage a change in the worktree.
3. Configure fake-opencode with background git loop mock.
4. Run gen-commit-msg with `--dir` pointing at the worktree.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainDir := filepath.Join(req.TempDir, "main-repo")
	worktreeDir := filepath.Join(req.TempDir, "linked-worktree")
	InitGitRepoWithWorktree(t, mainDir, worktreeDir)

	req.GitDir = worktreeDir
	WriteFile(t, filepath.Join(worktreeDir, "session.go"), "package session\n// AutoGenerateSessionID\n")
	RunGit(t, worktreeDir, "add", "session.go")

	WriteMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_wt_race","llm_events":[{"type":"step_start"},{"type":"tool_call","tool":"bash","tool_input":{"command":"(lock=$(git rev-parse --git-path index.lock); touch \"$lock\") &"}},{"type":"message","text":"{\"title\": \"feat: worktree commit\", \"description\": \"Commit from linked worktree\"}"},{"type":"step_finish"}]}`)
	req.Commit = true
	return nil
}
```