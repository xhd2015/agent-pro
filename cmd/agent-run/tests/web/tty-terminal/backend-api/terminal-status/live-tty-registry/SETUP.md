# Scenario

**Feature**: live tty registry is advertised as attachable

```
codex-tty session + reachable codex-tty-registry/<id>.json -> GET /terminal -> available true
```

## Preconditions

- Registry `listen_addr` points to a reachable ptywrap-compatible server.

## Steps

1. Start fake ptywrap server.
2. Write matching `codex-tty` session and registry entry.
3. Fetch terminal status.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex-tty"
	req.SessionID = "live-terminal-session"
	req.RegistryTranscript = "live-terminal\n"
	listenAddr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, listenAddr)
	req.HTTPPath = terminalStatusPath(req.Runner, req.SessionID)
	return nil
}
```
