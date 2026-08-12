# Scenario

**Feature**: short open prompt stays inline (no `--prompt-file`)

```
BuildFollowUpCommand(Prompt="hello", Open, PromptSpillDir=tmp)
  -> line contains `--` and shell-quoted hello
  -> no --prompt-file
  -> PromptSpillDir has no spill files
```

## Steps

1. Set short Prompt=`hello` (well under 600 runes).
2. Inject `PromptSpillDir` under `d.DOCTEST_CASE/spill`.
3. Open profile defaults from root.

```go
import (
	"testing"
	"unicode/utf8"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-short-open"
	req.Prompt = fixtureShortPrompt // "hello"
	req.Open = true
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.PromptSpillDir = dir
	if utf8.RuneCountInString(req.Prompt) > promptFileSpillMinRunes {
		t.Fatalf("fixture short prompt must be ≤%d runes; got %d",
			promptFileSpillMinRunes, utf8.RuneCountInString(req.Prompt))
	}
	return nil
}
```
