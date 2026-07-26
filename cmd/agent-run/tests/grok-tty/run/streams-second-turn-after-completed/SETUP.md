# Scenario

**Bug**: integration — stdout shows turn 2 events after turn 1 `turn_completed` while fake TUI runs

```
fake TUI holds 6s
  -> turn 1 ACP lines + turn_completed seeded
  -> scheduled turn 2 assistant STREAM_TURN_TWO_MARKER after turn_completed
  -> stdout marker before fake TUI exits
```

## Steps

1. Seed temp `GROK_HOME` session with turn 1 user + assistant + `turn_completed`.
2. Fake TUI holds 6 seconds after banner.
3. `Mode=stream-probe` schedules turn 2 append 1.2s after run start.
4. Assert stdout receives `STREAM_TURN_TWO_MARKER` while PTY still running.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

const (
	streamTurnTwoMarker   = "STREAM_TURN_TWO_MARKER"
	streamTurnTwoGrokUUID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	turnOnePromptText     = "turn one integration prompt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = streamTurnTwoGrokUUID
	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, streamTurnTwoGrokUUID, turnOnePromptText,
		acpAgentMessageChunk("turn one assistant done"),
		acpTurnCompleted(),
	)
	appendGrokHomeEnv(req)

	req.GrokTTYCommand = fakeTUIHoldSeconds(6)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, turnOnePromptText)

	req.Mode = "stream-probe"
	req.StreamProbeSubstring = streamTurnTwoMarker
	req.StreamProbeTimeout = 15 * time.Second
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{
		{
			Delay: 1200 * time.Millisecond,
			Lines: []string{
				acpUserMessageChunk("turn two user follow-up"),
				acpAgentMessageChunk(streamTurnTwoMarker),
				acpTurnCompleted(),
			},
		},
	}
	return nil
}
```