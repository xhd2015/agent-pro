# Scenario

**Feature**: live tail of grok `updates.jsonl` streams formatted events before PTY exits

```
fake TUI holds 5s
  -> agent-run discovers temp GROK_HOME session
  -> appends tool + assistant ACP lines mid-run
  -> stdout shows STREAM_PROBE_LS_DONE before fake TUI exits
```

## Steps

1. Seed temp `GROK_HOME` session dir with matching `user_message_chunk`.
2. Fake TUI holds 5 seconds after banner.
3. `Mode=stream-probe` schedules mid-run `updates.jsonl` appends.
4. Assert stdout receives probe marker while process still running.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

const streamProbeMarker = "STREAM_PROBE_LS_DONE"
const streamProbeGrokUUID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = streamProbeGrokUUID
	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, streamProbeGrokUUID, "stream probe")
	appendGrokHomeEnv(req)

	req.GrokTTYCommand = fakeTUIHoldSeconds(5)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, "stream probe")

	req.Mode = "stream-probe"
	req.StreamProbeSubstring = streamProbeMarker
	req.StreamProbeTimeout = 12 * time.Second
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{
		{
			Delay: 800 * time.Millisecond,
			Lines: []string{
				acpToolCall("call_stream_probe", "execute", "ls"),
				acpToolCallUpdate("call_stream_probe", "completed", "agent\nagents"),
				acpAgentMessageChunk(streamProbeMarker),
			},
		},
	}
	return nil
}
```