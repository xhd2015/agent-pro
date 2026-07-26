# Scenario

**Feature**: subagent resolves session via agent_session_id alias

```
# only agent_session_id in meta; SessionID matches
findOrCreateFlatSession -> resume (not new); Run completes
```

## Preconditions

- `meta.json` has `agent_session_id` but no `explicit_session_id`.
- Seed `events.jsonl` with a marker line to detect truncation vs append.

## Steps

1. Create flat session dir with host meta and one pre-existing event line.
2. Run with `SessionID` equal to `agent_session_id`.

## Context

- Session id: `gen_layout_alias_test`
- Inner session id: `inner_alias_sess`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "alias-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}
	req.SessionID = "gen_layout_alias_test"
	writeFile(t, filepath.Join(req.SessionDir, "meta.json"), `{
  "id": "20260625-fixture-alias",
  "agent_session_id": "gen_layout_alias_test",
  "task_title": "alias title"
}`)
	writeFile(t, filepath.Join(req.SessionDir, "events.jsonl"), `{"type":"message","text":"seed event","agent_runner":"fake-codex"}`+"\n")

	req.Layout = subagent.SessionLayout{
		Dir:              req.SessionDir,
		MessagesPath:     "",
		QuestionsEnabled: false,
		ProgressEnabled:  false,
	}

	mockPath := filepath.Join(req.TempDir, "alias-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_alias_sess"))
	req.MockConfigPath = mockPath
	return nil
}```
