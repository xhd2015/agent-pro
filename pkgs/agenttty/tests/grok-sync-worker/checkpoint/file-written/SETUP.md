# Scenario

**Feature**: S4 — grok-sync.json written with monotonic updates_offset

```
worker processes appended lines
  -> grok-sync.json exists on disk
  -> updates_offset matches EOF of processed content
```

## Steps

1. Start worker on empty file.
2. Append one complete turn.
3. Assert checkpoint file and offset monotonicity.

```go
import (
	"testing"
	"time"
)

const checkpointProbeUser = "checkpoint-file-probe"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InitialLines = nil
	req.AppendSchedules = []AppendSchedule{
		{
			Delay: 200 * time.Millisecond,
			Lines: []string{
				acpUserMessageChunk(checkpointProbeUser),
				acpAgentMessageChunk("checkpoint-reply"),
				acpTurnCompleted(),
			},
		},
	}
	req.HoldAfterSchedule = 1000 * time.Millisecond
	return nil
}
```
