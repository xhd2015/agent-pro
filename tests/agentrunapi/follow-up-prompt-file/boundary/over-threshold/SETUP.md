# Scenario

**Feature**: 601 runes crosses threshold → spill + `--prompt-file`

```
Prompt = 601×'a', Open, PromptSpillDir=tmp
  -> --prompt-file=<abs under spill>
  -> spill content == Prompt
  -> body not embedded on line
```

## Steps

1. Prompt = `strings.Repeat("a", 601)` (`promptFileSpillMinRunes + 1`).
2. Inject spill dir under `d.DOCTEST_CASE/spill`.

```go
import (
	"testing"
	"unicode/utf8"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-boundary-over"
	req.Prompt = longASCIIPrompt(promptFileSpillMinRunes + 1) // 601
	req.Open = true
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.PromptSpillDir = dir
	n := utf8.RuneCountInString(req.Prompt)
	if n != promptFileSpillMinRunes+1 {
		t.Fatalf("over-threshold fixture must be %d runes; got %d",
			promptFileSpillMinRunes+1, n)
	}
	return nil
}
```
