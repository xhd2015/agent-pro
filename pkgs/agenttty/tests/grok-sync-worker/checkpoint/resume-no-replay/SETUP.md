# Scenario

**Feature**: S3 — checkpoint resume does not replay turn 1 events

```
turn 1 on disk -> worker processes -> StopGrokSync (checkpoint saved)
  -> EnsureGrokSync restart from checkpoint
  -> append turn 2 only
  -> turn 1 user text count unchanged (no replay)
```

## Steps

1. Seed turn 1 in `InitialLines`.
2. Set `StopAfterTurn=1`, `RestartAfterStop=true`.
3. Schedule turn 2 append via `PostRestartAppendSchedules`.

```go
import (
	"testing"
	"time"
)

const (
	resumeTurnOneUser = "resume-turn-one-user"
	resumeTurnTwoUser = "resume-turn-two-user"
	resumeTurnTwoMarker = "RESUME_TURN_TWO_MARKER"
)

func Setup(t *testing.T, req *Request) error {
	req.InitialLines = []string{
		acpUserMessageChunk(resumeTurnOneUser),
		acpAgentMessageChunk("resume-turn-one-assistant"),
		acpTurnCompleted(),
	}
	req.StopAfterTurn = 1
	req.RestartAfterStop = true
	req.PostRestartAppendSchedules = []AppendSchedule{
		{
			Delay: 200 * time.Millisecond,
			Lines: []string{
				acpUserMessageChunk(resumeTurnTwoUser),
				acpAgentMessageChunk(resumeTurnTwoMarker),
				acpTurnCompleted(),
			},
		},
	}
	req.HoldAfterSchedule = 1000 * time.Millisecond
	return nil
}
```
