# Scenario

**Feature**: SSE `after` byte offset skips events already on disk before subscribe

```
POST create -> wait finished -> SSE after=file size at subscribe -> no duplicate user line
```

## Preconditions

- `ReadEvents` offset semantics match `events.jsonl` byte positions.

## Steps

1. Create session and wait until `finished`.
2. Record end offset of `events.jsonl` (or use store read offset).
3. Open SSE with `after=<offset>`; expect no events (or only post-subscribe writes).

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WebPort = 0
	req.SessionRunner = "fake-codex"
	req.CreatePrompt = "offset skip"
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID
	waitForSessionStatus(t, req, req.SessionRunner, sessionID, "finished", 30*time.Second)

	eventsPath := filepath.Join(req.Home, "sessions", sessionID, "events.jsonl")
	info, err := os.Stat(eventsPath)
	if err != nil {
		t.Fatalf("stat events: %v", err)
	}
	req.SSEAfterOffset = info.Size()
	req.SSEMaxWait = 3 * time.Second
	return nil
}
```