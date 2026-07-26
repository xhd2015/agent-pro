# Scenario

**Feature**: WaitReady times out when never ready

```
StatusFn always not-ready + short Timeout -> error (timeout/ready)
```

## Steps

1. Hold not-ready; Timeout ~40ms; PollInterval ~5ms.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-wait-timeout"
	req.ReadyTimeout = 40 * time.Millisecond
	req.ReadyPollInterval = 5 * time.Millisecond
	req.StatusPollSeq = nil
	req.StatusPollHold = statusNotReadyFixture()
	return nil
}
```
