# Scenario

**Bug**: grok session stderr diagnostics must appear before stdout streaming begins

```
fake TUI holds 5s
  -> discovery finds temp GROK_HOME session
  -> stderr grok session + updates lines before stdout stream marker
  -> mid-run updates.jsonl append emits STREAM_ORDER_MARKER on stdout
```

## Steps

1. Seed temp `GROK_HOME` session with matching prompt.
2. `Mode=stream-probe` schedules mid-run assistant marker on `updates.jsonl`.
3. Assert stderr grok session lines precede first stdout stream marker.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

const orderProbeMarker = "STREAM_ORDER_MARKER"
const orderProbeGrokUUID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = orderProbeGrokUUID
	prompt := "stream order probe"
	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, orderProbeGrokUUID, prompt)
	appendGrokHomeEnv(req)

	req.GrokTTYCommand = fakeTUIHoldSeconds(5)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, prompt)

	req.Mode = "stream-probe"
	req.StreamProbeSubstring = orderProbeMarker
	req.StreamProbeTimeout = 12 * time.Second
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{
		{
			Delay: 800 * time.Millisecond,
			Lines: []string{
				acpToolCall("call_order_probe", "execute", "ls"),
				acpToolCallUpdate("call_order_probe", "completed", "agent\nagents"),
				acpAgentMessageChunk(orderProbeMarker),
			},
		},
	}
	return nil
}
```