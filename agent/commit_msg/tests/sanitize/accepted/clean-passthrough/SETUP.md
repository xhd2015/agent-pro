# Scenario

**Feature**: clean JSON agent responses pass through sanitize unchanged

```
# agent returns normal {"title","description"} with no anti-patterns
fake-opencode -> parse -> sanitize no-op -> same formatted message
```

## Preconditions
- Fixture `clean_json_unchanged` (regression / false-positive guard).

## Steps
1. Stage a change.
2. Mock clean JSON agent text.
3. Run without `--commit`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	StageRepoWithChange(t, req)
	WriteMockAgentText(t, req, "sess_clean", ReadAntiPatternIn(t, "clean_json_unchanged"))
	req.Commit = false
	req.Operation = "clean_json_unchanged"
	return nil
}
```
