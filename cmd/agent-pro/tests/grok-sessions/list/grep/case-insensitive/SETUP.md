# Scenario

**Feature**: grep matching is case-insensitive for literal substrings

```
# content has mixed-case "LocalAgentXyz"; pattern is all-lowercase "localagentxyz"
writeGrokSession(title) + writeChatHistory(user) -> ListWithGrep("localagentxyz")

# session matches; hit snippets preserve original content casing
```

## Preconditions

- Pattern case differs from stored text (ASCII case fold).
- Both summary title and chat content use the mixed-case form.

## Steps

1. Set `req.Grep = "localagentxyz"` (all lowercase), `req.Limit = 10`, `req.Color = "never"`.
2. Write session title `Ship LocalAgentXyz CLI`.
3. Write chat user line containing `LocalAgentXyz packaging`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Grep = "localagentxyz"
	req.Limit = 10
	req.Color = "never"

	summaryPath := writeGrokSession(t, req.GrokHome,
		"01900017-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-case",
		"Ship LocalAgentXyz CLI")
	writeChatHistory(t, sessionDirOf(summaryPath), []chatHistoryMsg{
		{Type: "user", Content: "document LocalAgentXyz packaging next to agent-pro"},
	})
	return nil
}
```
