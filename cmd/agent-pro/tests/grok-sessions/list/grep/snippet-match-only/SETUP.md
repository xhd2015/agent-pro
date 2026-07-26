# Scenario

**Feature**: when the match itself is ≥1024 runes, the snippet is the first 1024 runes of the match only

```
# chat assistant content is a single long token (≥1024 runes) that is also the grep pattern
writeGrokSession + writeChatHistory(long match body) -> ListWithGrep(same long pattern)

# snippet = first 1024 runes of the match; no side context from outside the match
[]SessionMatch.Hits[0].Snippet -> rune count ≤1024, prefix of match text
```

## Preconditions

- Match length `M ≥ 1024` runes (pattern and field content share the same long token).
- Field may be exactly the match, or the match alone is long enough that windowing
  uses the match-only rule (first 1024 runes of the match).
- Title does not contain the pattern characters in a shorter accidental hit
  (use a short unrelated title; pattern is a pure run of `M` runes).
- No requirement for leading/trailing `...` under the match-only branch (snippet
  is truncated match text, not side-context ellipsis).

## Steps

1. Build `longMatch = 1100×'X'` (ASCII; rune count equals byte length).
2. Set `req.Grep = longMatch`, `req.Limit = 10`, `req.Color = "never"`.
3. Write one session with a short title without `X`-only match concerns; chat
   `assistant` content is exactly `longMatch` (1100 runes).

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	longMatch := strings.Repeat("X", 1100)
	req.Grep = longMatch
	req.Limit = 10
	req.Color = "never"

	summaryPath := writeGrokSession(t, req.GrokHome,
		"01900022-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-snippet-match-only",
		"Match-only window session")
	writeChatHistory(t, sessionDirOf(summaryPath), []chatHistoryMsg{
		{Type: "assistant", Content: longMatch},
	})
	return nil
}
```
