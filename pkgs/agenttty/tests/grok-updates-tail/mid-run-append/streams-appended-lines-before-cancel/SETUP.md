# Scenario

**Feature**: lines appended while tail is watching stream before context cancel

```
seed user chunk only
  -> tail starts at offset 0
  -> delayed tool + assistant append (MID_RUN_APPEND_MARKER)
  -> marker in EventTexts before ctx cancel
```

## Steps

1. Seed minimal turn-1 user chunk (no assistant yet).
2. Schedule tool + assistant append 250ms after tail start.
3. Assert mid-run marker appears in collected events.

```go
import (
	"testing"
	"time"
)

const (
	midRunSeedUserText = "mid run seed user"
	midRunAppendMarker = "MID_RUN_APPEND_MARKER"
)

func Setup(t *testing.T, req *Request) error {
	req.StartOffset = 0
	req.InitialLines = []string{
		acpUserMessageChunk(midRunSeedUserText),
	}
	req.TailStartDelay = 200 * time.Millisecond
	req.AppendSchedules = []AppendSchedule{
		{
			Delay: 250 * time.Millisecond,
			Lines: []string{
				acpToolCall("call_mid_run", "execute", "ls"),
				acpToolCallUpdate("call_mid_run", "completed", "agent\nagents"),
				acpAgentMessageChunk(midRunAppendMarker),
			},
		},
	}
	req.HoldAfterSchedule = 700 * time.Millisecond
	return nil
}
```