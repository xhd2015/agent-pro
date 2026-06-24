# Scenario

Generate a commit message from staged changes using fake-opencode mock events.

## Preconditions
- A git repo has staged text changes.
- fake-opencode mock returns a JSON commit message.

## Steps
1. Initialize git repo and stage a change.
2. Run gen-commit-msg without `--commit`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.GitDir = filepath.Join(req.TempDir, "repo")
	InitGitRepo(t, req.GitDir)
	WriteFile(t, filepath.Join(req.GitDir, "feature.go"), "package main\n")
	RunGit(t, req.GitDir, "add", "feature.go")

	WriteMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_commit","llm_events":[{"type":"step_start"},{"type":"message","text":"{\"title\": \"feat: add feature\", \"description\": \"Implement feature X\"}"},{"type":"step_finish"}]}`)

	req.Commit = false
	return nil
}
```