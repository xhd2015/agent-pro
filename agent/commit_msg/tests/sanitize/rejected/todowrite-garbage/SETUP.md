# Scenario

**Feature**: todowrite / tool-noise agent text is rejected (hard fail, no commit)

```
# agent returns: `todowrite` the commit message.`todowrite` completed.
fake-opencode -> sanitize rejects as unusable -> non-zero
--commit must not create a commit (HEAD unchanged)
```

## Preconditions
- Fixture `todowrite_garbage` with `.want_err`.
- Policy: hard failure; no LLM auto-retry.

## Steps
1. Stage a change; record HEAD.
2. Mock garbage agent text.
3. Run with `--commit`.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	StageRepoWithChange(t, req)
	head, err := execGitOutput(req.GitDir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("rev-parse HEAD before run: %v", err)
	}
	// Reuse WorktreeDir to carry prior HEAD for Assert (unused otherwise in this leaf).
	req.WorktreeDir = strings.TrimSpace(head)
	WriteMockAgentText(t, req, "sess_todowrite", ReadAntiPatternIn(t, "todowrite_garbage"))
	req.Commit = true
	req.Operation = "todowrite_garbage"
	return nil
}
```
