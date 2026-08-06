# Scenario

**Feature**: known session with two user turns yields two prompts with wire times

```
# two user messages separated by assistant/turn
user@14:30 "hello world" + user@14:45 "second prompt"
  -> Prompts -> 2 UserPrompts with matching timestamps and text
```

## Preconditions

- Session id `idKnownTwo` discoverable via summary.json.
- Wire timestamps are top-level ms on each user_message_chunk.
- Times: 30m and 15m before fixedNow (2026-07-03 15:00:00Z).

## Steps

1. Write session with two user chunks and intervening assistant/turn_completed.
2. Call Prompts.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "single"
	req.SessionID = idKnownTwo
	t1 := atFixed(-30 * time.Minute) // 14:30:00Z
	t2 := atFixed(-15 * time.Minute) // 14:45:00Z
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idKnownTwo,
		Title:        "two prompts fixture",
		LastActiveAt: t2,
		Updates: updatesJSONL(
			userChunkAt("hello world", t1),
			assistantChunk("hi there"),
			turnCompleted(),
			userChunkAt("second prompt", t2),
			assistantChunk("ok"),
			turnCompleted(),
		),
	})
	return nil
}
```
