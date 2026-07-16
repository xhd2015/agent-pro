# Scenario

**Feature**: `sessions … --print` — formatted session trace by bare id or
`--grok-session-id`

```
agent-run sessions <session_id> --print -> GetSession + events.jsonl -> FormatState -> stdout
agent-run sessions --print --grok-session-id UUID -> meta-only resolve (grok|grok-tty)
# running: tail events + poll meta.status until not running
# bare session id only for positional; reject runner/id (Q5)
# --grok-session-id mutually exclusive with positional <session_id>
```

## Preconditions

- Positional ref is a single bare `session_id` token (no `/`).
- `--print` is required when printing (with positional or `--grok-session-id`).
- Grok lookup matches `meta.runner` exactly `grok` or `grok-tty` and
  `meta.runner_session_id` (trim-space equality).

## Steps

1. Grouping `Setup` sets `req.SessionRunner` and `req.SessionID` when leaves omit them.
2. Leaf `Setup` seeds storage and builds `req.Args` as `sessions <id> --print`,
   `sessions --print --grok-session-id …`, or CLI error variants.
3. `Run` executes the CLI against `AGENT_RUN_HOME`.

```go
import (
	"testing"
)

func printSessionArgs(sessionID string, extra ...string) []string {
	args := []string{"sessions", sessionID, "--print"}
	return append(args, extra...)
}

// printGrokSessionArgs builds sessions --print --grok-session-id ID (flag order flexible).
func printGrokSessionArgs(grokSessionID string, extra ...string) []string {
	args := []string{"sessions", "--print", "--grok-session-id", grokSessionID}
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
