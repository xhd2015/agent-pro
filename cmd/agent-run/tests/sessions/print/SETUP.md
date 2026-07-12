# Scenario

**Feature**: `sessions <session_id> --print` — formatted session trace

```
agent-run sessions <session_id> --print -> GetSession + events.jsonl -> FormatState -> stdout
# running: tail events + poll meta.status until not running
# bare session id only; reject runner/id (Q5)
```

## Preconditions

- Positional ref is a single bare `session_id` token (no `/`).
- `--print` is required when the positional is present.

## Steps

1. Grouping `Setup` sets `req.SessionRunner` and `req.SessionID` when leaves omit them.
2. Leaf `Setup` seeds storage and builds `req.Args` as `sessions <id> --print` (or CLI error variants).
3. `Run` executes the CLI against `AGENT_RUN_HOME`.

```go
import (
	"testing"
)

func printSessionArgs(sessionID string, extra ...string) []string {
	args := []string{"sessions", sessionID, "--print"}
	return append(args, extra...)
}

func Setup(t *testing.T, req *Request) error {
	if req.SessionRunner == "" {
		req.SessionRunner = printRunner
	}
	if req.SessionID == "" {
		req.SessionID = printSessionID
	}
	return nil
}
```
