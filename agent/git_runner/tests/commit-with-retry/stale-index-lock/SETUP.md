# Scenario

**Feature**: CommitWithRetry removes stale index.lock and completes the commit

```
staged change + leftover index.lock
  -> CommitWithRetry (RemoveStaleIndexLock each attempt)
  -> commit succeeds with expected subject
```

## Preconditions

- Temp git repo with a staged file.
- Stale `index.lock` will be created in `Run` via `Interference=stale-lock`.

## Steps

1. Init repo, stage `next.txt`.
2. Set message and `Interference=stale-lock`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoDir = filepath.Join(t.TempDir(), "repo")
	initGitRepo(t, req.RepoDir)
	stageNewFile(t, req.RepoDir, "next.txt", "world\n")
	req.Message = "feat: after stale lock"
	req.MaxAttempts = 5
	req.Interference = "stale-lock"
	req.NoVerify = false
	return nil
}
```
