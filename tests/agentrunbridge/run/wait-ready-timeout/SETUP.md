# Scenario

**Feature**: wait-ready times out when status never ready

```
launch -> status(not-ready)* until ReadyTimeout -> error (timeout / not banner+sendable)
```

## Preconditions

- Always not-ready via `StatusPollHold`.
- Short ReadyTimeout so leaf finishes quickly.

## Steps

1. WaitReady true; ReadyTimeout 50ms; hold not-ready forever.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "wait timeout"
	req.SessionID = "sess-wait-to"
	req.AgentRunner = "grok-tty"
	req.AutoSendOrResume = true
	req.NewTerminal = true
	req.Open = true
	req.WaitReady = true
	req.ReadyTimeout = 50 * time.Millisecond
	req.ReadyPollInterval = 10 * time.Millisecond
	req.StatusPollHold = statusNotReadyFixture()
	return nil
}
```
