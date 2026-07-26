# Scenario

**Feature**: HostOwnsMeta with agent_session_id alias — no explicit_session_id or meta patch

```
# meta-alias scenario with HostOwnsMeta: resume via SessionID, append events, meta frozen
agent_session_id alias -> Run -> append events ; no explicit_session_id added
```

## Preconditions

- `meta.json` has `agent_session_id` only (no `explicit_session_id`, no `opencode_session_id`).
- Seed `events.jsonl` marker line.

## Steps

1. Create flat dir with host meta and seed event.
2. Snapshot meta bytes; run with matching `SessionID`.

## Context

- Session id: `gen_layout_alias_host_test`
- Inner session id: `inner_alias_host_sess`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "alias-host-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}
	req.SessionID = "gen_layout_alias_host_test"
	metaPath := filepath.Join(req.SessionDir, "meta.json")
	writeFile(t, metaPath, `{
  "id": "20260625-fixture-alias-host",
  "agent_session_id": "gen_layout_alias_host_test",
  "task_title": "alias host title"
}`)
	writeFile(t, filepath.Join(req.SessionDir, "events.jsonl"), `{"type":"message","text":"seed event","agent_runner":"fake-codex"}`+"\n")

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
	mockPath := filepath.Join(req.TempDir, "alias-host-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_alias_host_sess"))
	req.MockConfigPath = mockPath
	return nil
}```
