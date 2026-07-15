# Scenario

**Feature**: legitimate inner code-span backticks are preserved

```
# title contains intentional inner ticks: feat: add `--open` flag
fake-opencode -> sanitize unwraps only outer pairs -> inner ` kept
```

## Preconditions
- Fixture `legitimate_inner_backticks`.
- Rule: never strip legitimate inner backticks (only matching outer wrappers).

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
	WriteMockAgentText(t, req, "sess_inner_ticks", ReadAntiPatternIn(t, "legitimate_inner_backticks"))
	req.Commit = false
	req.Operation = "legitimate_inner_backticks"
	return nil
}
```
