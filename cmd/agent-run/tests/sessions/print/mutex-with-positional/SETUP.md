# Scenario

**Feature**: `--grok-session-id` and positional `<session_id>` are mutually exclusive
for `sessions --print`

```
seed session
  -> agent-run sessions <id> --print --grok-session-id UUID
  -> exit 1; mutually exclusive
```

## Preconditions

- Session exists so failure is mutex validation, not missing meta.

## Steps

1. Seed finished session with matching UUID.
2. Run print with both positional session id and `--grok-session-id`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	printMutexSessionID  = "print_gsid_mutex_s1"
	printMutexRunnerUUID = "550e8400-e29b-41d4-a716-446655440921"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	store := openAgentStore(t, req)
	req.SessionID = printMutexSessionID
	req.SessionRunner = "grok-tty"
	seedSessionMetaRunnerSessionID(t, store, "grok-tty", printMutexSessionID, "finished", printMutexRunnerUUID)
	req.Args = []string{
		"sessions",
		printMutexSessionID,
		"--print",
		"--grok-session-id", printMutexRunnerUUID,
	}
	return nil
}
```
