# Scenario

**Feature**: empty open prompt keeps current behavior (no `--prompt-file`)

```
BuildFollowUpCommand(Prompt="", Open, PromptSpillDir=tmp)
  -> current open behavior (`--` with empty body per existing rules)
  -> no --prompt-file
  -> no spill files
```

## Steps

1. Set Prompt to empty string (override any root default).
2. Inject spill dir; expect no writes.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-empty-open"
	req.Prompt = ""
	req.Open = true
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.PromptSpillDir = dir
	return nil
}
```
