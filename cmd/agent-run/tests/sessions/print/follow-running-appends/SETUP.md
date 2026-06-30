# Scenario

**Feature**: follow running session until status is no longer running

```
seed running + 1 event -> --print tails -> append 2nd event + status finished -> Session finished footer
```

## Preconditions

- `meta.status` is `running` when the CLI starts.
- A sidecar appends a second distinct event and flips status after ~500ms.

## Steps

1. Seed `fake-codex/follow_append` with status `running` and one message event.
2. Start CLI with extended timeout; sidecar appends `Follow-up appended line` and sets `finished`.

```go
import (
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

const followSessionID = "follow_append"

func Setup(t *testing.T, req *Request) error {
	store := openAgentStore(t, req)
	seedSessionMeta(t, store, printRunner, followSessionID, "running")
	appendAgentMessage(t, store, printRunner, followSessionID, "First running event")

	req.SessionID = followSessionID
	req.Args = printSessionArgs(printRunner, followSessionID)
	req.ExecTimeout = 15 * time.Second

	home := req.Home
	runner := printRunner
	sid := followSessionID
	req.Sidecar = func() {
		time.Sleep(500 * time.Millisecond)
		s, err := agentstorage.NewFileStore(home)
		if err != nil {
			return
		}
		_ = s.AppendEvent(runner, sid, types.AgentEvent{
			Type: types.ActionMessage,
			Role: "assistant",
			Text: "Follow-up appended line",
		})
		_ = s.UpdateSessionStatus(runner, sid, "finished")
	}
	return nil
}
```