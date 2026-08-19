# Scenario

**Bug**: C2 — `sessions --print` follow must deliver appends after status becomes `finished`

```
seed running session -> --print enters follow (WatchEvents)
  -> sidecar sets status finished then appends marker
  -> stdout must include appended text (status must not stop watch)
```

## Steps

1. Seed **running** session with one assistant line (follow mode active).
2. Run `sessions grok-tty/<id> --print` with 12s timeout.
3. Sidecar: wait 600ms, flip status to `finished`, append `CHAT_TAIL_CLI_FOLLOW_MARKER`.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const chatTailCLIFollowMarker = "CHAT_TAIL_CLI_FOLLOW_MARKER"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "cli-follow-delivers-after-finished"
	req.Runner = "grok-tty"
	req.SessionID = "chat_tail_cli_follow"
	seedRunningSessionForFollow(t, req, req.Runner, req.SessionID)
	// Flat layout: sessions CLI expects bare session_id (not runner/id).
	req.CLIArgs = []string{"sessions", req.SessionID, "--print"}
	req.ExecTimeout = 12 * time.Second
	runner := req.Runner
	sid := req.SessionID
	req.Sidecar = func() {
		time.Sleep(200 * time.Millisecond)
		markSessionFinished(t, req, runner, sid)
		// Append after WatchEvents status poll (500ms) has observed finished.
		time.Sleep(1500 * time.Millisecond)
		appendSessionEvent(t, req, runner, sid, chatTailCLIFollowMarker)
	}
	return nil
}
```