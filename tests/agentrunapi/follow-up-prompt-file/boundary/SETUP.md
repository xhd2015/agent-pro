# Scenario

**Feature**: threshold boundary — 600 runes inline; 601 runes spill

```
runeCount(TrimSpace(Prompt)) ≤ 600 → inline (no --prompt-file)
runeCount(TrimSpace(Prompt)) > 600 → spill + --prompt-file
```

## Steps

1. Grouping documents the locked threshold `promptFileSpillMinRunes = 600`.
2. Leaves set ASCII pad length exactly at / over the threshold.
3. Always inject `PromptSpillDir` under `d.DOCTEST_CASE`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Leaves set SessionID / Prompt / PromptSpillDir.
	if !req.Open && !req.Detach {
		req.Open = true
	}
	return nil
}
```
