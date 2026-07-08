# Scenario

**Bug**: A1 — rapid follow-ups must not duplicate user events (PRIMARY duplicate guard)

```
agentsync.EnsureGrokSync starts single worker
  -> turn 1 (hello?) appended + turn_completed
  -> turn 2 (what did I say?) appended within overlap window
  -> exactly one user message per distinct prompt
```

## Steps

1. Start worker on empty `updates.jsonl`.
2. Schedule turn 1 complete (user + assistant + `turn_completed`) at 200ms.
3. Schedule turn 2 (user + assistant + `turn_completed`) at 400ms.
4. Assert one user line per prompt text.

```go
import (
	"testing"
	"time"
)

const (
	syncPromptA = "hello?"
	syncPromptB = "what did I say?"
	syncReplyA  = "reply-to-hello"
	syncReplyB  = "reply-to-recall"
)

func Setup(t *testing.T, req *Request) error {
	req.InitialLines = nil
	req.AppendSchedules = []AppendSchedule{
		{
			Delay: 200 * time.Millisecond,
			Lines: []string{
				acpUserMessageChunk(syncPromptA),
				acpAgentMessageChunk(syncReplyA),
				acpTurnCompleted(),
			},
		},
		{
			Delay: 200 * time.Millisecond,
			Lines: []string{
				acpUserMessageChunk(syncPromptB),
				acpAgentMessageChunk(syncReplyB),
				acpTurnCompleted(),
			},
		},
	}
	req.HoldAfterSchedule = 1200 * time.Millisecond
	return nil
}
```