# Scenario

**Feature**: CLI sessions --print still tails running session after WatchEvents extraction

```
seed running + event -> --print tails -> append second event + finish -> stdout includes appended line
```

## Steps

1. Seed running session with first event.
2. Run `sessions <runner>/<id> --print` with sidecar append.

```go
import (
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "fake-codex"
	req.SessionID = "watch-events-print"
	seedRunningSessionForPrint(t, req, req.Runner, req.SessionID)
	req.CLIArgs = []string{"sessions", req.Runner + "/" + req.SessionID, "--print"}
	req.ExecTimeout = 15 * time.Second
	home := req.Home
	runner := req.Runner
	sid := req.SessionID
	req.Sidecar = func() {
		time.Sleep(500 * time.Millisecond)
		store, err := agentstorage.NewFileStore(home)
		if err != nil {
			return
		}
		_ = store.AppendEvent(runner, sid, types.AgentEvent{
			Type: types.ActionMessage,
			Role: "assistant",
			Text: "WatchEvents appended line",
		})
		_ = store.UpdateSessionStatus(runner, sid, "finished")
	}
	req.Mode = "cli"
	return nil
}
```
