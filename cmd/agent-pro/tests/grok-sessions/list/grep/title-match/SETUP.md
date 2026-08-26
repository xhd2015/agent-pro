# Scenario

**Feature**: grep pattern matching only generated_title lists the session with a title hit line

```
# one session title contains GREP_TITLE_TOKEN; another session does not
writeGrokSession(title with token) + writeGrokSession(unrelated) -> ListWithGrep

# only matching session returned; hit: summary.json:1:title: ...
[]SessionMatch + FormatListTableWithHits(color=never)
```

## Preconditions

- Pattern is a case-insensitive literal substring of `generated_title` only.
- No `chat_history.jsonl` (or none containing the pattern).
- A second session exists without the pattern so filter omission is observable.

## Steps

1. Set `req.Grep = []string{"GREP_TITLE_TOKEN"}`, `req.Limit = 10`, `req.Color = "never"`.
2. Write matching session with title containing the token.
3. Write non-matching session with a different title.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Grep = []string{"GREP_TITLE_TOKEN"}
	req.Limit = 10
	req.Color = "never"

	writeGrokSession(t, req.GrokHome,
		"01900010-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-title-match",
		"Implement GREP_TITLE_TOKEN feature")
	writeGrokSession(t, req.GrokHome,
		"01900010-bbbb-7bbb-bbbb-bbbbbbbbbbbb",
		"2026-07-03T14:45:00.000Z",
		"/tmp/grep-title-other",
		"Unrelated session about widgets")
	return nil
}
```
