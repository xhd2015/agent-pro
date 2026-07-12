# Scenario

**Feature**: multiple hits under one session print ordered indented lines (summary then chat)

```
# pattern GREP_MULTI_TOKEN in title, session_summary, user, and assistant lines
writeGrokSessionOpts + writeChatHistory -> ListWithGrep

# hit order: summary title, session_summary, then chat_history lines in file order
FormatListTableWithHits -> multiple "  file:line:part: snippet" lines
```

## Preconditions

- Single session with several hits; total hits ≤ 5 so no overflow line.
- Summary field hits come before chat hits.
- Within summary: title-ish fields before cwd/model (title, then session_summary).

## Steps

1. Set `req.Grep = "GREP_MULTI_TOKEN"`, `req.Limit = 10`, `req.Color = "never"`.
2. Write one session with title + session_summary containing the token.
3. Write chat history: system (no token), user (token), assistant (token).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Grep = "GREP_MULTI_TOKEN"
	req.Limit = 10
	req.Color = "never"

	summaryPath := writeGrokSessionOpts(t, req.GrokHome,
		"01900012-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-multi",
		"Design GREP_MULTI_TOKEN for list",
		grokSessionOpts{
			NumChatMessages: 3,
			SessionSummary:  "Session covers GREP_MULTI_TOKEN end-to-end",
		})
	writeChatHistory(t, sessionDirOf(summaryPath), []chatHistoryMsg{
		{Type: "system", Content: "system prompt without the marker"},
		{Type: "user", Content: "add GREP_MULTI_TOKEN search under sessions"},
		{Type: "assistant", Content: "Implementing GREP_MULTI_TOKEN filter now"},
	})
	return nil
}
```
