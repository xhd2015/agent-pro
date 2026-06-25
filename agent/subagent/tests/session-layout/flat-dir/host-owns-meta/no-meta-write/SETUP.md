# Scenario

**Feature**: HostOwnsMeta prevents subagent from writing meta.json

```
# host meta bytes fixed before Run; subagent writes events only
host meta.json (frozen) -> subagent.Run -> events.jsonl ; meta bytes unchanged
```

## Preconditions

- Pre-created task-hub-like `meta.json` with `agent_session_id`.
- `HostOwnsMeta` true; callback wired but does not persist to disk in this leaf.

## Steps

1. Create flat session dir and write host `meta.json`.
2. Snapshot exact meta bytes before `Run`.
3. Run with `SessionID` matching `agent_session_id`.

## Context

- Inner session id from mock: `inner_host_no_write`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "host-no-write")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}
	req.SessionID = "gen_layout_host_no_write"
	hostMeta := `{
  "id": "20260625-fixture-host-no-write",
  "task_title": "host owns meta",
  "agent_session_id": "gen_layout_host_no_write",
  "project_name": "fixture-a",
  "created_at": "2026-06-25T10:00:00Z"
}`
	metaPath := filepath.Join(req.SessionDir, "meta.json")
	writeFile(t, metaPath, hostMeta)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatal(err)
	}
	req.MetaBytesBeforeRun = data

	req.Layout = subagent.SessionLayout{
		Dir:              req.SessionDir,
		MessagesPath:     "",
		QuestionsEnabled: false,
		ProgressEnabled:  false,
	}
	mockPath := filepath.Join(req.TempDir, "host-no-write-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_host_no_write"))
	req.MockConfigPath = mockPath
	return nil
}```
