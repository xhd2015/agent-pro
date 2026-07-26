# Scenario

**Feature**: web codex-tty POST stores terminal_session_id from HeadlessRun

```
POST codex-tty -> meta.terminal_session_id populated with registry id
```

## Steps

1. Create web codex-tty session with fake TUI.
2. Read session detail for terminal_session_id.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	createWebCodexTTYSession(t, req, "store terminal id")
	waitForSessionStatus(t, req, req.Runner, req.ChatSessionID, "finished", 60*time.Second)
	req.TerminalSessionID = waitForTerminalSessionID(t, req, req.Runner, req.ChatSessionID, 10*time.Second)
	req.Mode = "http"
	req.HTTPPath = "/api/agent-run/sessions/" + req.Runner + "/" + req.ChatSessionID
	return nil
}
```
