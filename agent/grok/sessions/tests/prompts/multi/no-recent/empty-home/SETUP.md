# Scenario

**Feature**: empty grok home / no sessions returns empty list

```
# sessions/ exists but has no session dirs
ListPrompts -> [] ; format-list -> "No user prompts found\n"
```

## Preconditions

- Root Setup created empty `sessions/` only.
- Op `format-list` so empty friendly message is checked.

## Steps

1. Do not write any sessions.
2. Call ListPrompts + FormatPromptsListText.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-list"
	req.RecentSet = false
	req.LimitSet = false
	return nil
}
```
