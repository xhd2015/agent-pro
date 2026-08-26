# Scenario

**Feature**: list sessions filtered by case-insensitive literal grep over session JSON

```
# grep patterns filter sessions that have ≥1 hit in summary.json or chat_history.jsonl
# multiple patterns → AND on the same field/line
sessions.ListWithGrep(grokHome, limit, patterns) -> []SessionMatch (newest matches first)

# each match row may be followed by indented hit lines (capped at 5) with optional color
[]SessionMatch -> FormatListTableWithHits(now, colorMode) -> table + hit details
```

## Preconditions

- Descendants set `req.Grep` as `[]string` (except `no-grep-no-hits`, which leaves it empty).
- `req.Color` is `never` or `always` for deterministic color tests; default empty → harness uses `never`.
- Hit line shape: two-space indent + `<file>:<line>:<part>: <snippet>`.
- Hit order: summary.json field hits first, then chat_history.jsonl in file order.
- Part for `generated_title` is `title`; chat parts follow message `type`.
- Snippet window (after whitespace collapse): at most 1024 runes; ASCII `...`
  (3 runes) on truncated sides; ~50/50 before/after first pattern; match ≥1024 → first
  1024 runes of the match only. `MatchStart`/`MatchLen` are byte offsets into
  the final snippet for the first pattern. Short fields must not gain false ellipsis.

## Steps

1. Ensure list operation (parent already sets `req.Operation = "list"`).
2. Leaf Setup writes fixtures and sets Grep / Color / Limit.
3. Run uses ListWithGrep + FormatListTableWithHits when `len(Grep) > 0`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Parent list Setup already set Operation=list and created sessions root.
	// Default color off for plain-output leaves; color leaves override.
	if req.Color == "" {
		req.Color = "never"
	}
	return nil
}
```
