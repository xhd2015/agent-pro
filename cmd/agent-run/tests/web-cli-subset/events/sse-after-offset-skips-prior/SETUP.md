# Scenario

**Feature**: SSE `after` byte offset skips events already on disk

```
POST session -> wait finished -> SSE after=EOF offset -> zero replay
```

## Steps

1. Create fake-codex session and wait until finished.
2. Subscribe SSE at end-of-file byte offset.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	startAgentRunWeb(t, req)
	req.Runner = "fake-codex"
	req.Prompt = "offset skip cli parity"
	sessionID, _, _ := postCreateSession(t, req, req.Runner, req.Prompt)
	req.SessionID = sessionID
	waitForSessionStatus(t, req, req.Runner, sessionID, "finished", 45*time.Second)
	eventsPath := filepath.Join(req.Home, "sessions", req.Runner, sessionID, "events.jsonl")
	info, err := os.Stat(eventsPath)
	if err != nil {
		t.Fatalf("stat events: %v", err)
	}
	req.SSEAfterOffset = info.Size()
	req.SSEMaxWait = 3 * time.Second
	req.Mode = "sse"
	return nil
}
```
