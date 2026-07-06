# Scenario

`--commit --no-verify` must pass `--no-verify` to `git commit`, skipping a failing pre-commit hook.

## Preconditions
- A git repo has a pre-commit hook that always exits 1.
- Staged text changes and fake-opencode mock return a commit message.
- Without `--no-verify`, `git commit` would fail at the hook (contrast documented in ASSERT, not a separate leaf).

## Steps
1. Initialize git repo with failing pre-commit hook via `InitGitRepoWithFailingPreCommitHook`.
2. Stage a text file change.
3. Configure fake-opencode mock events.
4. Run gen-commit-msg with `--commit` and `--no-verify`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.GitDir = filepath.Join(req.TempDir, "repo")
	InitGitRepoWithFailingPreCommitHook(t, req.GitDir)
	WriteFile(t, filepath.Join(req.GitDir, "feature.go"), "package main\n")
	RunGit(t, req.GitDir, "add", "feature.go")

	WriteMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_no_verify","llm_events":[{"type":"step_start"},{"type":"message","text":"{\"title\": \"feat: skip hooks\", \"description\": \"Commit with --no-verify\"}"},{"type":"step_finish"}]}`)

	req.Commit = true
	req.NoVerify = true
	return nil
}
```