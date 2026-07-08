# Scenario

**Bug**: `WatchEvents` must tail until ctx cancelled — session status is not a gate

```
seed finished session + initial event
  -> WatchEvents(ctx, after=EOF)
  -> append new assistant line while ctx alive
  -> onLine receives appended text
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentevents` exports `WatchEvents`.
- Tests use isolated `t.TempDir()` as `AGENT_RUN_HOME`.
- No agent-run CLI required.

## Steps

1. Root `Setup` sets defaults for timing fields.
2. Leaf `Setup` seeds finished session and append marker text.
3. `Run` starts watch goroutine, schedules append, cancels ctx after hold.
4. Leaf `Assert` checks appended line was received.

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentevents"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

const defaultWatchAppendMarker = "WATCHEVENTS_FINISHED_APPEND_MARKER"

func Setup(t *testing.T, req *Request) error {
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	if req.AppendDelay <= 0 {
		req.AppendDelay = 400 * time.Millisecond
	}
	if req.WatchHold <= 0 {
		req.WatchHold = 2 * time.Second
	}
	if req.AppendText == "" {
		req.AppendText = defaultWatchAppendMarker
	}
	return nil
}

func seedFinishedSession(t *testing.T, req *Request) agentstorage.Store {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	meta := agentstorage.SessionMeta{
		Runner:    req.Runner,
		SessionID: req.SessionID,
		Status:    "finished",
		Workspace: req.TempDir,
	}
	if err := store.CreateSession(req.Runner, req.SessionID, meta); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := store.UpdateSessionStatus(req.Runner, req.SessionID, "finished"); err != nil {
		t.Fatalf("UpdateSessionStatus: %v", err)
	}
	if err := store.AppendEvent(req.Runner, req.SessionID, types.AgentEvent{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: "Initial finished event",
	}); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
	return store
}

func eventsTailOffset(t *testing.T, store agentstorage.Store, runner, sessionID string) int64 {
	t.Helper()
	_, offset, err := store.ReadEvents(runner, sessionID, 0)
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}
	return offset
}

func runWatchEventsProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	store := seedFinishedSession(t, req)
	if req.AfterOffset <= 0 {
		req.AfterOffset = eventsTailOffset(t, store, req.Runner, req.SessionID)
	}

	var mu sync.Mutex
	var received []string
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		err := agentevents.WatchEvents(ctx, store, req.Runner, req.SessionID, req.AfterOffset, func(line string) error {
			mu.Lock()
			received = append(received, line)
			mu.Unlock()
			return nil
		})
		done <- err
	}()

	time.Sleep(req.AppendDelay)
	if err := store.AppendEvent(req.Runner, req.SessionID, types.AgentEvent{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: req.AppendText,
	}); err != nil {
		return nil, fmt.Errorf("append while watching: %w", err)
	}

	time.Sleep(req.WatchHold)
	cancel()
	watchErr := <-done

	mu.Lock()
	lines := append([]string(nil), received...)
	mu.Unlock()

	resp := &Response{
		ReceivedLines: lines,
		WatchErr:      watchErr,
	}
	for _, line := range lines {
		var ev map[string]any
		if json.Unmarshal([]byte(line), &ev) == nil {
			if text, _ := ev["text"].(string); text != "" {
				resp.ReceivedTexts = append(resp.ReceivedTexts, text)
			}
		}
		if strings.Contains(line, req.AppendText) {
			resp.GotAppended = true
		}
	}
	return resp, nil
}
```