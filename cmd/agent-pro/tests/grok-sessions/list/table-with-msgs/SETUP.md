# Scenario

**Feature**: list table includes MSGS column from num_chat_messages

```
# one session with num_chat_messages=42
writeGrokSessionOpts -> sessions.List -> FormatListTable(now)

# table header has MSGS; row shows 42
terminal table text
```

## Preconditions

- `req.Now` is fixed at `2026-07-03T15:00:00.000Z` by root Setup.
- `summary.json` includes `num_chat_messages`.

## Steps

1. Create one session with `num_chat_messages=42` and a non-empty title.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Limit = 10
	writeGrokSessionOpts(t, req.GrokHome,
		"01900005-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/msgs-project",
		"Session with messages",
		grokSessionOpts{NumChatMessages: 42})
	return nil
}
```