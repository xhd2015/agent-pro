# Scenario

**Feature**: non-transient hook failures fail immediately (no infinite / pointless retry)

```
failing pre-commit hook + staged change
  -> CommitWithRetry(maxAttempts=5)
  -> returns error; output is not classified as transient
```

## Preconditions

- Repo has a real pre-commit hook that exits 1 (hooksPath not nulled).
- Staged change ready for commit.

## Steps

1. Init repo with failing husky-style pre-commit.
2. Stage `hooked.txt`.
3. Set `Interference=hook-fail` (no extra Run-side injection).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoDir = filepath.Join(t.TempDir(), "repo")
	initGitRepoWithFailingHook(t, req.RepoDir)
	stageNewFile(t, req.RepoDir, "hooked.txt", "hook me\n")
	req.Message = "feat: should not commit past hook"
	req.MaxAttempts = 5
	req.Interference = "hook-fail"
	req.NoVerify = false
	return nil
}
```
