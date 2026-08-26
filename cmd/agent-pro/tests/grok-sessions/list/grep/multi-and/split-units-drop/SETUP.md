# Scenario

**Feature**: patterns on different lines do not satisfy AND — session dropped

```
# line1 has AND_ONLY_A; line2 has AND_ONLY_B; no line has both
--grep AND_ONLY_A --grep AND_ONLY_B -> No sessions found
```

## Preconditions

- Single session; patterns never co-occur in the same field/line.
- Distinct from same-unit-keep (which also has a split other session as a
  secondary check).

## Steps

1. Set both patterns.
2. Write one session with each pattern on a separate chat line.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Grep = []string{"AND_ONLY_A", "AND_ONLY_B"}
	req.Limit = 10
	req.Color = "never"

	summary := writeGrokSession(t, req.GrokHome,
		"01900031-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-and-split",
		"Title without tokens")
	writeChatHistory(t, sessionDirOf(summary), []chatHistoryMsg{
		{Type: "user", Content: "message with AND_ONLY_A alone"},
		{Type: "assistant", Content: "reply with AND_ONLY_B alone"},
	})
	return nil
}
```
