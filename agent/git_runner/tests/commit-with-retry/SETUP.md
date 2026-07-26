# Scenario

**Feature**: CommitWithRetry recovers from transient index errors and fails fast on permanent ones

```
temp git repo + interference
  -> CommitWithRetry(dir, msg, attempts, noVerify)
doctest <- commit subject on success, or non-retryable error
```

## Preconditions

- `git` is on PATH.
- Leaf initializes `RepoDir`, stages a change, sets `Interference` and `Message`.

## Steps

1. Set `req.Mode = "commit-with-retry"`.
2. Default `MaxAttempts` to 5 when unset.
3. Leaf prepares repo and interference; `Run` executes recovery scenario.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Mode != "" && req.Mode != "commit-with-retry" {
		return fmt.Errorf("commit-with-retry subtree requires Mode=commit-with-retry, got %q", req.Mode)
	}
	req.Mode = "commit-with-retry"
	ensureGitAvailable(t)
	if req.MaxAttempts < 1 {
		req.MaxAttempts = 5
	}
	return nil
}
```
