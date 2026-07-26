# Scenario

**Feature**: CLI attach relays stdin to PTY and stdout from PTY via AttachRelay

```
background stub-tty -> agent-run attach <terminal-id> + stdin -> marker echoed
```

## Steps

1. Start background stub-tty; capture terminal session id.
2. Run `agent-run attach` with stdin marker via CLI mode.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.TerminalSessionID = startStubTTYBackground(t, req)
	req.CLIArgs = []string{"attach", req.TerminalSessionID}
	req.CLIStdin = "CLI_ATTACH_MARKER\n"
	req.Mode = "cli"
	req.ExecTimeout = 15 * time.Second
	return nil
}
```
