# Scenario

**Feature**: subagent merges agent fields into pre-existing host-owned meta.json

```
# task-hub meta.json exists before Run; subagent updates opencode_session_id only
host meta.json -> subagent.Run -> same id/task_title + new opencode_session_id
```

## Preconditions

- `meta.json` pre-created with task-hub-like fields and `agent_session_id` alias.

## Steps

1. Create flat session dir.
2. Write foreign `meta.json` before subagent run.
3. Run with `SessionID` matching `agent_session_id`.

## Context

- Inner session id from mock: `inner_merged_sess`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "merged-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}
	req.SessionID = "gen_layout_merged_test"
	req.PreCreateMeta = `{
  "id": "20260625-fixture-a-merged",
  "task_title": "keep this title",
  "agent_session_id": "gen_layout_merged_test",
  "project_name": "fixture-a",
  "created_at": "2026-06-25T10:00:00Z"
}`
	writeFile(t, filepath.Join(req.SessionDir, "meta.json"), req.PreCreateMeta)
	req.Layout = subagent.SessionLayout{
		Dir:              req.SessionDir,
		MessagesPath:     "",
		QuestionsEnabled: false,
		ProgressEnabled:  false,
	}
	mockPath := filepath.Join(req.TempDir, "merged-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_merged_sess"))
	req.MockConfigPath = mockPath
	return nil
}
```
