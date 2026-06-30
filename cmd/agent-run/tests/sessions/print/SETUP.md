# Scenario

**Feature**: `sessions <runner>/<session_id> --print` — formatted session trace

```
agent-run sessions <runner>/<id> --print -> GetSession + events.jsonl -> FormatState -> stdout
# running: tail events + poll meta.status until not running
```

## Preconditions

- Positional ref is a single token `runner/session_id`.
- `--print` is required when the positional is present.

## Steps

1. Grouping `Setup` sets `req.SessionRunner` and `req.SessionID` when leaves omit them.
2. Leaf `Setup` seeds storage and builds `req.Args` as `sessions <runner>/<id> --print` (or CLI error variants).
3. `Run` executes the CLI against `AGENT_RUN_HOME`.

```go
import (
	"fmt"
	"testing"
)

func printSessionArgs(runner, sessionID string, extra ...string) []string {
	ref := fmt.Sprintf("%s/%s", runner, sessionID)
	args := []string{"sessions", ref, "--print"}
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