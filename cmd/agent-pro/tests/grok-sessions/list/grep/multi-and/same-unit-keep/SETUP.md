# Scenario

**Feature**: both patterns on the same chat line keep the session and show that hit

```
# user line: "AND_ALPHA … AND_BETA …"
--grep AND_ALPHA --grep AND_BETA -> keep; hit line contains both tokens
```

## Preconditions

- Matching session has one user line containing both tokens.
- Non-matching session has neither token (or only one elsewhere is not needed;
  this leaf focuses on the positive same-unit path).

## Steps

1. Set Grep to AND_ALPHA and AND_BETA.
2. Write matching session with both tokens on one user line.
3. Write other session without both tokens together.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Grep = []string{"AND_ALPHA", "AND_BETA"}
	req.Limit = 10
	req.Color = "never"

	matchSummary := writeGrokSession(t, req.GrokHome,
		"01900030-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-and-same",
		"Plain title no tokens")
	writeChatHistory(t, sessionDirOf(matchSummary), []chatHistoryMsg{
		{Type: "user", Content: "please handle AND_ALPHA with AND_BETA together"},
		{Type: "assistant", Content: "ok"},
	})

	otherSummary := writeGrokSession(t, req.GrokHome,
		"01900030-bbbb-7bbb-bbbb-bbbbbbbbbbbb",
		"2026-07-03T14:40:00.000Z",
		"/tmp/grep-and-other",
		"Other session")
	writeChatHistory(t, sessionDirOf(otherSummary), []chatHistoryMsg{
		{Type: "user", Content: "only AND_ALPHA here"},
		{Type: "assistant", Content: "only AND_BETA here"},
	})
	return nil
}
```
