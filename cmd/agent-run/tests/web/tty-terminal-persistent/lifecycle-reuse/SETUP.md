# Scenario

**Bug**: web tty lifecycle must reuse the mapped backend PTY

```
web chat web_* -> terminal_session_id session-1 -> live PTY
navigation or follow-up -> same terminal_session_id session-1
```

## Preconditions

- A mapped live PTY already exists for the web chat.
- Reuse is required while that mapped PTY is reachable.

## Steps

1. Descendant setup writes mapped session metadata and live registry.
2. `Run` performs navigation-style reattach or follow-up message probe.
3. Assertion checks the same mapped PTY remains in use.

## Context

- A future replacement policy may allocate a new PTY only after the old PTY is
  stale; that is not this branch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RegistryTranscript = "reused-terminal-session-1\n"
	listenAddr := startMappedPtywrap(t, req)
	writeMappedSessionFixture(t, req)
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, listenAddr)
	return nil
}
```
