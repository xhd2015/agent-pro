# Scenario

**Feature**: single `git commit -m "..."` wrapper yields the inner title

```
# agent returns `git commit -m "feat: ..."` (often outer-backticked)
fake-opencode -> sanitize extracts first -m argument -> title only
```

## Preconditions
- Fixture `git_commit_m_wrapper`.

## Steps
1. Stage a change.
2. Mock agent text from fixture.
3. Run without `--commit`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	StageRepoWithChange(t, req)
	WriteMockAgentText(t, req, "sess_git_m", ReadAntiPatternIn(t, "git_commit_m_wrapper"))
	req.Commit = false
	req.Operation = "git_commit_m_wrapper"
	return nil
}
```
