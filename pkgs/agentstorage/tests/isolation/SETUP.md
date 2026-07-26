# Scenario

**Feature**: all durable writes remain under `AGENT_RUN_HOME`

```
CreateSession + SaveConfig + AppendEvent + AppendMessage
-> scan home tree -> every file path has home prefix
# session files under sessions/<session_id>/ only
```

## Preconditions

- Store must never create files outside the resolved home directory.
- Test performs multiple write types in one run to exercise all code paths.

## Steps

1. Set `req.Operation = "isolation"`.
2. Leaf Setup sets runner, session, config, and message payload.
3. `Run` performs representative writes then scans the home tree.
4. Leaf `Assert` calls `AssertHomeOnly` on `Response.FilesWritten`.

## Context

- `Response.FilesWritten` is populated by walking the home directory after writes.
- Parent `AssertHomeOnly` rejects any path not under the home prefix.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "isolation"
	req.Action = "writes_under_home"
	if req.Runner == "" {
		req.Runner = "fake-opencode"
	}
	if req.SessionID == "" {
		req.SessionID = "sess_iso"
	}
	if req.Config.DefaultAgentRunner == "" {
		req.Config = agentstorage.Config{
			DefaultAgentRunner: "fake-opencode",
			DefaultModel:       "test-model",
		}
	}
	return nil
}
```
