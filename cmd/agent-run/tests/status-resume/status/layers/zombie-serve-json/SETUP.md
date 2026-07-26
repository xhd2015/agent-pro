# Scenario

**Bug**: `status --json` after zombie serve `/exit` must expose `runner.exited: true` and `resume.ready: true` while process/terminal stay live

```
seed zombie serve (alive PID + reachable + exit scrollback)
  -> agent-run status --json test-zombie-json-s1
  -> JSON process.alive, terminal.reachable, runner.exited=true, resume.ready=true
```

## Steps

1. Seed zombie serve fixture (bound + exit markers + live registry).
2. Run `status --json`.
3. Mode `status-json` parses stdout.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "test-zombie-json-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440224"
	req.TerminalSessionID = "term-zombie-json-1"
	req.InitialPrompt = "zombie json after exit"
	seedZombieServeAfterExit(t, req)
	req.Mode = "status-json"
	req.Args = []string{"status", "--json", req.SessionID}
	return nil
}
```
