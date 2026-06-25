# Scenario

**Feature**: OnAgentComplete receives inner runner session id from mock

```
# fake-codex returns session_id; callback delivers InnerSessionID to host
subagent.Run -> mock inner id -> OnAgentComplete(InnerSessionID, AgentRunner)
```

## Preconditions

- Flat dir with minimal host meta (no `opencode_session_id`).
- `HostOwnsMeta` true.

## Steps

1. Create flat session dir and host `meta.json`.
2. Run with mock `session_id` = `inner_callback_sess`.

## Context

- Session id: `gen_layout_callback_test`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "callback-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}
	req.SessionID = "gen_layout_callback_test"
	writeFile(t, filepath.Join(req.SessionDir, "meta.json"), `{
  "id": "20260625-fixture-callback",
  "agent_session_id": "gen_layout_callback_test",
  "task_title": "callback test"
}`)
	req.Layout = subagent.SessionLayout{
		Dir:              req.SessionDir,
		MessagesPath:     "",
		QuestionsEnabled: false,
		ProgressEnabled:  false,
	}
	mockPath := filepath.Join(req.TempDir, "callback-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_callback_sess"))
	req.MockConfigPath = mockPath
	req.AgentRunner = "fake-codex"
	return nil
}```
