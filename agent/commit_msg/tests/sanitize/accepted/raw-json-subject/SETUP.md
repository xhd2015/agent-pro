# Scenario

**Feature**: raw JSON object agent text becomes formatted title + description

```
# agent returns a bare {"title","description"} object as the step text
fake-opencode -> parse/sanitize -> formatted message (never print raw JSON as subject)
```

## Preconditions
- Fixture `json_raw_subject`.

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
	WriteMockAgentText(t, req, "sess_raw_json", ReadAntiPatternIn(t, "json_raw_subject"))
	req.Commit = false
	req.Operation = "json_raw_subject"
	return nil
}
```
