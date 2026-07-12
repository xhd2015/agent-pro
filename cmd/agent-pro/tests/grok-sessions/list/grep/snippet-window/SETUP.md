# Scenario

**Feature**: long matching field is windowed around the match with a 1024-rune cap and `...` ellipsis

```
# tool_result text is much longer than 1024 runes; first CI match sits in the middle
writeGrokSession + writeChatHistory(long tool_result) -> ListWithGrep("SNIP_WIN_NEEDLE")

# hit snippet is collapsed whitespace, ≤1024 runes, includes NEEDLE, leading+trailing "..."
[]SessionMatch -> FormatListTableWithHits(color=never) -> indented hit line
```

## Preconditions

- Pattern `SNIP_WIN_NEEDLE` appears only once, in a long `tool_result` chat field.
- Raw field text is well over 1024 runes after whitespace collapse (e.g. ~600 prefix
  + pattern + ~600 suffix, plus newline runs that collapse to single spaces).
- Title and other fields do not contain the pattern (single hit).
- Snippet algorithm: whitespace collapse first; then window ≤1024 runes including
  ASCII `...` (3 runes) on each truncated side; ~50/50 before/after the match.

## Steps

1. Set `req.Grep = "SNIP_WIN_NEEDLE"`, `req.Limit = 10`, `req.Color = "never"`.
2. Write one session with a short title that does not contain the pattern.
3. Write `chat_history.jsonl` with a single `tool_result` line whose content is
   `600×'a'` + newlines + `SNIP_WIN_NEEDLE` + newlines + `600×'b'`.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Grep = "SNIP_WIN_NEEDLE"
	req.Limit = 10
	req.Color = "never"

	// Long tool_result: mid-field match; newlines must collapse before windowing.
	longBody := strings.Repeat("a", 600) + "\n\n" + "SNIP_WIN_NEEDLE" + "\n\n" + strings.Repeat("b", 600)

	summaryPath := writeGrokSession(t, req.GrokHome,
		"01900020-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-snippet-window",
		"Long tool dump session")
	writeChatHistory(t, sessionDirOf(summaryPath), []chatHistoryMsg{
		{Type: "tool_result", Content: longBody},
	})
	return nil
}
```
