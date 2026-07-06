# Scenario

**Bug**: second turn streams after `turn_completed` while tail context is alive

```
updates.jsonl: turn1 user + assistant + turn_completed (pre-seeded)
  -> TailUpdatesFromOffset startOffset=0
  -> append turn2 user + TURN_TWO_TAIL_MARKER assistant
  -> marker in EventTexts before ctx cancel
```

## Steps

1. Pre-seed turn 1 with `turn_completed`.
2. Schedule turn 2 append 300ms after tail start.
3. Hold tail 600ms after last append; cancel and assert turn 2 marker.

```go
import (
	"testing"
	"time"
)

const (
	turnOneUserText      = "first turn user prompt"
	turnOneAssistantText = "first turn assistant reply"
	turnTwoUserText      = "second turn user prompt"
	turnTwoTailMarker    = "TURN_TWO_TAIL_MARKER"
)

func Setup(t *testing.T, req *Request) error {
	req.StartOffset = 0
	req.InitialLines = []string{
		acpUserMessageChunk(turnOneUserText),
		acpAgentMessageChunk(turnOneAssistantText),
		acpTurnCompleted(),
	}
	req.TailStartDelay = 200 * time.Millisecond
	req.AppendSchedules = []AppendSchedule{
		{
			Delay: 300 * time.Millisecond,
			Lines: []string{
				acpUserMessageChunk(turnTwoUserText),
				acpAgentMessageChunk(turnTwoTailMarker),
			},
		},
	}
	req.HoldAfterSchedule = 800 * time.Millisecond
	return nil
}
```