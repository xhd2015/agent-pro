# Scenario

**Feature**: info shows (untitled) for empty-title bootstrap session

```
# empty generated_title with num_chat_messages=1, no updates.jsonl
writeGrokSessionOpts -> sessions.Info -> FormatInfoText(now)

# Title line uses (untitled); message count still shown
terminal key-value text
```

## Preconditions

- Empty title correlates with init-only sessions (0–1 chat messages, no updates.jsonl).
- `num_chat_messages=1` must appear in output.

## Steps

1. Write a session with empty title and `num_chat_messages=1`.
2. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const untitledSessionID = "019f283a-cccc-7ccc-cccc-cccccccccccc"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = untitledSessionID
	writeGrokSessionOpts(t, req.GrokHome, untitledSessionID,
		"2026-07-03T14:45:00.000Z",
		"/tmp/grok-untitled",
		"",
		grokSessionOpts{
			NumMessages:     0,
			NumChatMessages: 1,
		})
	return nil
}
```