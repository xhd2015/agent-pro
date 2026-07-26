# Scenario

**Feature**: grep pattern that matches nothing yields empty list message

```
# sessions exist on disk but none contain GREP_NO_MATCH_TOKEN
writeGrokSession x2 -> ListWithGrep(pattern)

# empty matches -> FormatListTableWithHits -> "No sessions found"
```

## Preconditions

- At least one session fixture exists so empty result is from filtering, not empty home.
- Pattern string does not appear in any summary or chat content.

## Steps

1. Set `req.Grep = "GREP_NO_MATCH_TOKEN"`, `req.Limit = 10`, `req.Color = "never"`.
2. Write two sessions with unrelated titles and chat text.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Grep = "GREP_NO_MATCH_TOKEN"
	req.Limit = 10
	req.Color = "never"

	s1 := writeGrokSession(t, req.GrokHome,
		"01900014-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-nomatch-a",
		"Alpha widgets session")
	writeChatHistory(t, sessionDirOf(s1), []chatHistoryMsg{
		{Type: "user", Content: "talk about alpha widgets only"},
	})

	s2 := writeGrokSession(t, req.GrokHome,
		"01900014-bbbb-7bbb-bbbb-bbbbbbbbbbbb",
		"2026-07-03T14:20:00.000Z",
		"/tmp/grep-nomatch-b",
		"Beta gadgets session")
	writeChatHistory(t, sessionDirOf(s2), []chatHistoryMsg{
		{Type: "assistant", Content: "beta gadgets are fine"},
	})
	return nil
}
```
