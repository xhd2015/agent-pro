# Scenario

**Feature**: wait-ready succeeds after not-ready then ready poll

```
launch -> status(not-ready) -> status(ready) -> Run ok
```

## Preconditions

- `WaitReady=true`.
- Short poll interval; status sequence length 2.
- Session id used in status poll argv by implementer.

## Steps

1. Enable WaitReady; script StatusPollSeq not-ready then ready.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "wait ok"
	req.SessionID = "sess-wait-ok"
	req.AgentRunner = "grok-tty"
	req.AutoSendOrResume = true
	req.NewTerminal = true
	req.Open = true
	req.WaitReady = true
	req.ReadyTimeout = 2 * time.Second
	req.ReadyPollInterval = 10 * time.Millisecond
	req.StatusPollSeq = []string{
		statusNotReadyFixture(),
		statusReadyFixture(),
	}
	return nil
}
```
