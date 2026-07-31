# Scenario

**Feature**: no registry TTY path yields zero iTerm candidates

```
# session exists; registry pid present; tree only ??/blank; iTerm list empty of matches
FocusSession -> error (none found); FocusITerm never called
```

## Preconditions

- Session meta + registry seed so resolve reaches tree walk.
- Procs under serve PID have no real TTYs.
- ITermRefs do not match any collected TTY (empty list).

## Steps

1. Seed session + registry; inject empty/blank-only tree and empty iTerm list.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-none"
	req.TermID = "term-none"
	req.RootPID = 400
	req.RegistryPID = 400
	req.SeedSession = true
	req.SeedRegistry = true
	req.Procs = []agentrunapi.ProcRow{
		{PID: 400, PPID: 1, TTY: "??", Cmd: "agent-run serve"},
		{PID: 401, PPID: 400, TTY: "", Cmd: "child"},
	}
	req.ITermRefs = nil
	return nil
}
```
