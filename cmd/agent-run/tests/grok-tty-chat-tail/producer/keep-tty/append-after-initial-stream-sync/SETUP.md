# Scenario

**Bug**: P2 — completion append after initial batch synced (`tailState.streamed` race)

```
pre-seed user + think + pending tool_call (synced quickly)
  -> 2.5s later: tool_call_update + agent_message + turn_completed
  -> marker must still appear despite early streamed sync
```

## Steps

1. Fixed session id `chat_tail_p2`.
2. Use longer completion delay (2.5s) so pending tool_call is visible in `events.jsonl` first.
3. Assert `HasPendingToolFirst` before assistant marker.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "append-after-initial-stream-sync"
	req.SessionID = "chat_tail_p2"
	req.CompletionDelay = 2500 * time.Millisecond
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{{
		Delay: req.CompletionDelay,
		Lines: completionAppendLines(),
	}}
	return nil
}
```