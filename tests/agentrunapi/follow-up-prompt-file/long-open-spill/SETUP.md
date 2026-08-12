# Scenario

**Feature**: long open prompt (>600 runes) auto-spills to `--prompt-file`

```
BuildFollowUpCommand(
  Prompt=601×"字", Open, PromptSpillDir=tmp, SessionID=sid)
  -> --prompt-file=<abs path under spill dir>
  -> line does NOT embed the long body after --
  -> spill file content == TrimSpace(Prompt)
```

## Steps

1. Build Prompt = `strings.Repeat("字", 601)` (601 UTF-8 runes; proves rune
   count, not byte length alone).
2. Inject empty `PromptSpillDir` under `d.DOCTEST_CASE/spill`.
3. Open profile.

```go
import (
	"testing"
	"unicode/utf8"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-long-spill"
	req.Prompt = longCJKPrompt(promptFileSpillMinRunes + 1) // 601 字
	req.Open = true
	dir, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.PromptSpillDir = dir
	n := utf8.RuneCountInString(req.Prompt)
	if n <= promptFileSpillMinRunes {
		t.Fatalf("long fixture must be >%d runes; got %d", promptFileSpillMinRunes, n)
	}
	return nil
}
```
