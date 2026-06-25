# Scenario

Reproduce auto-unstage failure when gen-commit-msg runs from a repo subdirectory while
git reports staged paths relative to the repository root.

## Preconditions
- Git repository root contains a nested `task-hub/` tree.
- A file named `task-hub` exists inside the nested directory (binary executable).
- Staged paths use the repo-root prefix, e.g. `task-hub/agents/do-task/PROMPT.md`.
- gen-commit-msg runs with `--dir` pointing at the nested subdirectory, not the repo root.

## Steps
1. Initialize a monorepo-style git repository at `repo/`.
2. Create `repo/task-hub/agents/do-task/PROMPT.md` and a `repo/task-hub/task-hub` binary file.
3. Stage `task-hub/agents/do-task/PROMPT.md` from the repository root.
4. Run gen-commit-msg with `--dir` set to `repo/task-hub/`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	repoRoot := filepath.Join(req.TempDir, "repo")
	subdir := filepath.Join(repoRoot, "task-hub")
	promptPath := filepath.Join(subdir, "agents", "do-task", "PROMPT.md")

	InitGitRepo(t, repoRoot)
	WriteFile(t, promptPath, "# do-task prompt\n")
	if err := os.WriteFile(filepath.Join(subdir, "task-hub"), []byte{0x7f, 0x45, 0x4c, 0x46}, 0755); err != nil {
		return err
	}
	RunGit(t, repoRoot, "add", "task-hub/agents/do-task/PROMPT.md")

	WriteMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_unstage","llm_events":[{"type":"step_start"},{"type":"message","text":"{\"title\": \"docs: update do-task prompt\", \"description\": \"Refresh agent prompt\"}"},{"type":"step_finish"}]}`)

	req.GitDir = subdir
	req.Commit = false
	return nil
}
```