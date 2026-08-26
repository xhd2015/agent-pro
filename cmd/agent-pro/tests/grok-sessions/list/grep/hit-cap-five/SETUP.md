# Scenario

**Feature**: more than five hits per session caps display at 5 plus overflow count

```
# one session with 8 chat user/assistant lines containing GREP_CAP_TOKEN (no title hit)
writeChatHistory x8 matching lines -> ListWithGrep

# FormatListTableWithHits shows 5 hit lines then "... and 3 more matches"
```

## Preconditions

- Total hits = 8, all in `chat_history.jsonl` (title/cwd without token).
- Display cap is exactly 5; remaining N = 3.
- Overflow line is indented with two spaces: `  ... and 3 more matches`.

## Steps

1. Set `req.Grep = []string{"GREP_CAP_TOKEN"}`, `req.Limit = 10`, `req.Color = "never"`.
2. Write one session with plain title.
3. Write 8 chat messages each containing the token (user and assistant alternating).

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Grep = []string{"GREP_CAP_TOKEN"}
	req.Limit = 10
	req.Color = "never"

	summaryPath := writeGrokSession(t, req.GrokHome,
		"01900013-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-cap",
		"Plain title without cap token")

	msgs := make([]chatHistoryMsg, 0, 8)
	for i := 1; i <= 8; i++ {
		typ := "user"
		if i%2 == 0 {
			typ = "assistant"
		}
		msgs = append(msgs, chatHistoryMsg{
			Type:    typ,
			Content: fmt.Sprintf("message %d mentions GREP_CAP_TOKEN explicitly", i),
		})
	}
	writeChatHistory(t, sessionDirOf(summaryPath), msgs)
	return nil
}
```
