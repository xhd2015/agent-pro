# Scenario

**Feature**: multi `-m` wrapper maps first `-m` to title and rest to description

```
# agent returns: git commit -m "title" -m "p1" -m "p2"
fake-opencode -> sanitize: title = 1st -m; body = remaining -m joined by \n\n
```

## Preconditions
- Fixture `git_commit_m_double`.
- Policy: never truncate; body paragraphs separated by blank lines.

## Steps
1. Stage a change.
2. Mock agent text from fixture.
3. Run without `--commit`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	StageRepoWithChange(t, req)
	WriteMockAgentText(t, req, "sess_git_m2", ReadAntiPatternIn(t, "git_commit_m_double"))
	req.Commit = false
	req.Operation = "git_commit_m_double"
	return nil
}
```
