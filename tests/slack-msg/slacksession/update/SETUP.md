# Scenario

**Feature**: session update sets workspace dir on map entry

```
# Caller updates session workspace
Caller -> slack-msg session update [--session-id ID] --dir PATH [--json]
  -> resolve session id
  -> require existing map entry + --dir (exists, is directory)
  -> store absolute dir; bump updated_at; preserve other fields
  -> human OK line or --json full entry
```

## Preconditions

- Action is `update` as second arg.
- `--dir` required; path must exist as a directory.
- Isolated HomeDir for sessions.json.

## Steps

1. Clear Slack env; isolate home when seeding.
2. Leaves create workspace dirs under WorkDir, seed map, set args.

## Context

- JSON success returns full updated map entry including `session_id` + `agent_session_id`.
- Errors: session id required / not found / nothing to update / dir does not exist / not a directory.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
const sessionUpdateFixtureID = "slack-channel-C0UPD44K5J6"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
