# Scenario

**Feature**: HostOwnsMeta preserves foreign meta (merged-meta scenario without subagent patch)

```
# same fixture as merged-meta; HostOwnsMeta=true → meta frozen, callback only
host meta (id, task_title, project_name) -> Run -> callback InnerSessionID ; meta unchanged
```

## Preconditions

- Pre-created task-hub-like `meta.json` matching `merged-meta` leaf fixture.
- `HostOwnsMeta` true.

## Steps

1. Write foreign meta before run; snapshot bytes.
2. Run with `SessionID` = `gen_layout_merged_host_test`.

## Context

- Inner session id: `inner_merged_host_sess`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "merged-host-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}
	req.SessionID = "gen_layout_merged_host_test"
	hostMeta := `{
  "id": "20260625-fixture-a-merged-host",
  "task_title": "keep this title",
  "agent_session_id": "gen_layout_merged_host_test",
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
	mockPath := filepath.Join(req.TempDir, "merged-host-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_merged_host_sess"))
	req.MockConfigPath = mockPath
	return nil
}```
