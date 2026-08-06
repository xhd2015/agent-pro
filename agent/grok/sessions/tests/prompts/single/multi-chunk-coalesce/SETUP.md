# Scenario

**Feature**: consecutive user_message_chunk lines coalesce into one prompt

```
# two chunks without intervening assistant
userChunk("hel") + userChunk("lo there")
  -> one UserPrompt Text="hello there"
  -> Timestamp = first chunk wire time
```

## Preconditions

- First chunk at fixedNow−20m; second chunk at fixedNow−19m (must not use second ts).
- No assistant/tool between the two user chunks.

## Steps

1. Write session with two consecutive user chunks then assistant.
2. Call Prompts.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "single"
	req.SessionID = idCoalesce
	tFirst := atFixed(-20 * time.Minute)
	tSecond := atFixed(-19 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idCoalesce,
		Title:        "coalesce fixture",
		LastActiveAt: tSecond,
		Updates: updatesJSONL(
			userChunkAt("hel", tFirst),
			userChunkAt("lo there", tSecond),
			assistantChunk("merged"),
			turnCompleted(),
		),
	})
	return nil
}
```
