# Scenario

**Feature**: JSON title wrapped in outer backticks is unwrapped before commit

```
# agent returns JSON with title = `feat: ...` (outer matching backticks)
fake-opencode -> parse JSON -> sanitize unwrap title -> stdout / git subject clean
```

## Preconditions
- Fixture `title_outer_backticks` in `testdata/anti_patterns/`.
- Run with `--commit` so git subject is also asserted.

## Steps
1. Stage a change in an isolated git repo.
2. Mock agent text from `title_outer_backticks.in`.
3. Run gen-commit-msg with `--commit`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	StageRepoWithChange(t, req)
	WriteMockAgentText(t, req, "sess_outer_ticks", ReadAntiPatternIn(t, "title_outer_backticks"))
	req.Commit = true
	req.Operation = "title_outer_backticks"
	return nil
}
```
