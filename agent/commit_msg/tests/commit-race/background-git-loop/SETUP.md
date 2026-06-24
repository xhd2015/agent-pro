# Scenario

Reproduce index.lock race: agent leaves background git loop running, then `--commit` runs.

## Preconditions
- A git repo has staged changes.
- fake-opencode mock starts a background git status loop before returning the commit message.
- gen-commit-msg runs with `--commit`.

## Steps
1. Initialize git repo and stage a change.
2. Configure fake-opencode to spawn `(while true; do git status; done) &` via bash tool_call.
3. Run gen-commit-msg with `--commit`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.GitDir = filepath.Join(req.TempDir, "repo")
	InitGitRepo(t, req.GitDir)
	WriteFile(t, filepath.Join(req.GitDir, "config.go"), "package config\n// session retry hint\n")
	RunGit(t, req.GitDir, "add", "config.go")

	WriteMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_race","llm_events":[{"type":"step_start"},{"type":"tool_call","tool":"bash","tool_input":{"command":"(lock=$(git rev-parse --git-path index.lock); touch \"$lock\") &"}},{"type":"message","text":"{\"title\": \"feat: session ID auto-generation\", \"description\": \"Add retry hint formatting\"}"},{"type":"step_finish"}]}`)
	req.Commit = true
	return nil
}
```