# Scenario

**Bug**: production `fatal: unable to write new index file` must be retryable after interference clears

```
Darwin: chflags uchg on .git/index
  -> single Commit fails with "unable to write new index file" (transient)
  -> chflags nouchg
  -> CommitWithRetry succeeds with expected subject
```

## Preconditions

- Darwin only (`chflags uchg`); non-Darwin leaves are skipped in `Run`.
- Temp repo with staged change; index must exist (after seed commit + add).

## Steps

1. Init repo, stage `b.txt`.
2. Set `Interference=uchg-then-clear` and recovery commit message.
3. `Run` forces first failure under uchg, clears flag, then `CommitWithRetry`.

## Context

- Same sequence as `script/debug/git-index-write-retry` scenarioImmutableThenRetry.
- Label `slow` optional; scenario is usually under a few seconds.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoDir = filepath.Join(t.TempDir(), "repo")
	initGitRepo(t, req.RepoDir)
	stageNewFile(t, req.RepoDir, "b.txt", "b\n")
	req.Message = "feat: recovered after index write fail"
	req.MaxAttempts = 5
	req.Interference = "uchg-then-clear"
	req.NoVerify = false
	return nil
}
```
