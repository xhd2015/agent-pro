# Scenario

**Feature**: list sessions filtered by case-insensitive literal grep over session JSON

```
# grep pattern filters sessions that have ≥1 hit in summary.json or chat_history.jsonl
sessions.ListWithGrep(grokHome, limit, pattern) -> []SessionMatch (newest matches first)

# each match row may be followed by indented hit lines (capped at 5) with optional color
[]SessionMatch -> FormatListTableWithHits(now, colorMode) -> table + hit details
```

## Preconditions

- Descendants set `req.Grep` (except `no-grep-no-hits`, which leaves it empty).
- `req.Color` is `never` or `always` for deterministic color tests; default empty → harness uses `never`.
- Hit line shape: two-space indent + `<file>:<line>:<part>: <snippet>`.
- Hit order: summary.json field hits first, then chat_history.jsonl in file order.
- Part for `generated_title` is `title`; chat parts follow message `type`.

## Steps

1. Ensure list operation (parent already sets `req.Operation = "list"`).
2. Leaf Setup writes fixtures and sets Grep / Color / Limit.
3. Run uses ListWithGrep + FormatListTableWithHits when Grep is non-empty.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Parent list Setup already set Operation=list and created sessions root.
	// Default color off for plain-output leaves; color leaves override.
	if req.Color == "" {
		req.Color = "never"
	}
	return nil
}
```
