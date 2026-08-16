# Scenario

**Feature**: session info shows one durable map entry

```
# Caller inspects a single session
Caller -> slack-msg session info [--session-id ID] [--json]
  -> resolve session id (flag or SLACK_MSG_SESSION_ID)
  -> load map entry + count messages.jsonl
  -> human key: value or --json object
```

## Preconditions

- Action is `info` as second arg.
- Session id required (flag or env).
- Isolated HomeDir for store paths.

## Steps

1. Clear Slack env; isolate home when seeding.
2. Leaves seed map (+ optional messages.jsonl) and set args/env.

## Context

- Human and JSON include `session_id`, `agent_session_id` (equal today),
  map fields, `message_count`, `session_dir`.
- Empty dir: human `-`; JSON `""`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
const sessionInfoFixtureID = "slack-channel-C0INFO44K5J6"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
