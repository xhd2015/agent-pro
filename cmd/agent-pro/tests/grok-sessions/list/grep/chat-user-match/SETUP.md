# Scenario

**Feature**: grep pattern matching only a chat user message lists that session with a user hit

```
# matching session: chat_history user line contains GREP_CHAT_USER_TOKEN; title does not
# non-matching session: different chat content
writeGrokSession + writeChatHistory -> ListWithGrep

# hit: chat_history.jsonl:<n>:user: ...; non-matching session omitted
```

## Preconditions

- Pattern appears only in a user message of the matching session.
- Titles and other chat lines must not contain the pattern.
- Physical jsonl line numbers are 1-based.

## Steps

1. Set `req.Grep = "GREP_CHAT_USER_TOKEN"`, `req.Limit = 10`, `req.Color = "never"`.
2. Matching session: title without token; chat line 1 system, line 2 user with token.
3. Non-matching session: chat user line without token.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Grep = "GREP_CHAT_USER_TOKEN"
	req.Limit = 10
	req.Color = "never"

	matchSummary := writeGrokSession(t, req.GrokHome,
		"01900011-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-chat-match",
		"Plain title without token")
	writeChatHistory(t, sessionDirOf(matchSummary), []chatHistoryMsg{
		{Type: "system", Content: "You are a helpful coding agent."},
		{Type: "user", Content: "please implement GREP_CHAT_USER_TOKEN in the list command"},
		{Type: "assistant", Content: "Sure, I will explore the codebase."},
	})

	otherSummary := writeGrokSession(t, req.GrokHome,
		"01900011-bbbb-7bbb-bbbb-bbbbbbbbbbbb",
		"2026-07-03T14:40:00.000Z",
		"/tmp/grep-chat-other",
		"Another plain title")
	writeChatHistory(t, sessionDirOf(otherSummary), []chatHistoryMsg{
		{Type: "user", Content: "discuss weather widgets only"},
		{Type: "assistant", Content: "Weather widgets are unrelated."},
	})
	return nil
}
```
