# Scenario

**Bug**: `resume --open "hello"` with zombie registry same id must not fail
`session id already in use` (reclaim + reuse terminal id)

```
seed zombie: session_id == terminal_session_id == test-open-v7
  (alive detached serve PID + reachable + exit scrollback)
  -> agent-run resume --open test-open-v7 "hello"
  -> reclaim zombie -> reserve same id -> open/inject path
  -> NOT: run: session id "test-open-v7" already in use
```

## Steps

1. Start detached sleep as zombie serve PID.
2. Seed bound+exited zombie fixture with SessionID == TerminalSessionID.
3. Run `resume --open <id> "hello"` with instant attach + fake TUI.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Mirror production bug: --session-id made terminal id == agent session id.
	req.SessionID = "test-open-v7"
	req.TerminalSessionID = "test-open-v7"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440701"
	req.InitialPrompt = "prior open turn"
	req.RegistryPID = startDetachedSleepPID(t)
	seedZombieServeAfterExit(t, req)

	req.OpenInstantAttach = true
	req.GrokTTYCommand = fakeTUIRespondHi()
	req.FollowupPrompt = "hello"
	req.Args = []string{
		"resume",
		"--open",
		req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	req.Mode = "read-meta"
	return nil
}
```
