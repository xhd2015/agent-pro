# Scenario

**Feature**: WaitReady succeeds on second poll

```
StatusFn: not-ready, then ready -> WaitReady nil
StatusPollCalls >= 2
```

## Steps

1. Short poll interval; sequence length 2.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-wait-ok"
	req.ReadyTimeout = 2 * time.Second
	req.ReadyPollInterval = 5 * time.Millisecond
	req.StatusPollSeq = []string{
		statusNotReadyFixture(),
		statusReadyFixture(),
	}
	return nil
}
```
