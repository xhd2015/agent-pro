# Scenario

**Feature**: S5 — turn_index restored from checkpoint into converter

```
pre-seed grok-sync.json with turn_index=1 (turn 0 completed)
  -> EnsureGrokSync resume
  -> append turn 2 user chunk
  -> user event has extensions.grok_session.turn_index == 1
```

## Steps

1. Write turn 0 content to `updates.jsonl` (already on disk).
2. Pre-seed checkpoint with `turn_index=1` and offset at EOF of turn 0.
3. Start worker; append turn 2 user line only.

```go
import (
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

const turnIndexRestoreUser = "turn-index-restore-user-prompt"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.TempDir = t.TempDir()
	req.UpdatesPath = filepath.Join(req.TempDir, "updates.jsonl")
	req.SessionDir = filepath.Join(req.TempDir, "session")
	turn0 := []string{
		acpUserMessageChunk("turn-zero-user"),
		acpAgentMessageChunk("turn-zero-assistant"),
		acpTurnCompleted(),
	}
	if err := writeUpdatesJSONL(req.UpdatesPath, turn0...); err != nil {
		return err
	}
	offset := updatesFileSize(t, req.UpdatesPath)
	req.PreCheckpoint = &agenttty.GrokSyncState{
		GrokSessionID: req.GrokSessionID,
		UpdatesPath:   req.UpdatesPath,
		UpdatesOffset: offset,
		TurnIndex:     1,
	}
	req.InitialLines = nil
	req.AppendSchedules = []AppendSchedule{
		{
			Delay: 300 * time.Millisecond,
			Lines: []string{
				acpUserMessageChunk(turnIndexRestoreUser),
				acpAgentMessageChunk("turn-index-restore-assistant"),
				acpTurnCompleted(),
			},
		},
	}
	req.HoldAfterSchedule = 1000 * time.Millisecond
	return nil
}
```
