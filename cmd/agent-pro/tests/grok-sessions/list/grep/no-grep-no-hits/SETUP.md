# Scenario

**Feature**: without grep, list table has no indented hit lines even when chat_history exists

```
# classic list path (Grep empty); session has chat_history with searchable text
writeGrokSession + writeChatHistory -> List + FormatListTable

# table rows only — no "  summary.json:" / "  chat_history.jsonl:" hit lines
```

## Preconditions

- `req.Grep` remains empty so Run uses classic `List` + `FormatListTable`.
- Chat history contains distinctive text that would match if grepped.
- Regression guard: grep hit formatting must not appear unless Grep is set.

## Steps

1. Leave `req.Grep` empty; set `req.Limit = 10`.
2. Write one session with title and multi-line chat_history.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Grep = ""
	req.Limit = 10
	// Color unused on classic path; leave default.

	summaryPath := writeGrokSessionOpts(t, req.GrokHome,
		"01900015-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-no-grep",
		"Session with history but no grep",
		grokSessionOpts{NumChatMessages: 2})
	writeChatHistory(t, sessionDirOf(summaryPath), []chatHistoryMsg{
		{Type: "user", Content: "WOULD_MATCH_IF_GREPPED unique phrase"},
		{Type: "assistant", Content: "reply also WOULD_MATCH_IF_GREPPED"},
	})
	return nil
}
```
