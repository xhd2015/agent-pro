# Scenario

**Feature**: exactly 600 runes stays inline (≤ threshold)

```
Prompt = 600×'a', Open, PromptSpillDir=tmp
  -> argv prompt after `--`
  -> no --prompt-file
  -> no spill files
```

## Steps

1. Prompt = `strings.Repeat("a", 600)` (exactly `promptFileSpillMinRunes`).
2. Inject spill dir under `d.DOCTEST_CASE/spill`.

```go
import (
	"testing"
	"unicode/utf8"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-boundary-at"
	req.Prompt = longASCIIPrompt(promptFileSpillMinRunes) // 600
	req.Open = true
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.PromptSpillDir = dir
	n := utf8.RuneCountInString(req.Prompt)
	if n != promptFileSpillMinRunes {
		t.Fatalf("at-threshold fixture must be exactly %d runes; got %d",
			promptFileSpillMinRunes, n)
	}
	return nil
}
```
