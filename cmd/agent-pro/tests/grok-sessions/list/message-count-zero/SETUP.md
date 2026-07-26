# Scenario

**Feature**: list table shows zero MSGS for empty-title bootstrap session

```
# empty-title session with num_chat_messages=0 (no real conversation)
writeGrokSessionOpts -> sessions.List -> FormatListTable(now)

# MSGS column shows 0
terminal table text
```

## Preconditions

- Empty `generated_title` mirrors Grok init-only sessions on disk.
- `num_chat_messages=0` is a valid display value.

## Steps

1. Create one session with empty title and `num_chat_messages=0`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Limit = 10
	writeGrokSessionOpts(t, req.GrokHome,
		"01900005-bbbb-7bbb-bbbb-bbbbbbbbbbbb",
		"2026-07-03T14:00:00.000Z",
		"/tmp/zero-msgs-project",
		"",
		grokSessionOpts{NumChatMessages: 0})
	return nil
}
```