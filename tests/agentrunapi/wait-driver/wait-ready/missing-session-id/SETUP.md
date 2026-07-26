# Scenario

**Feature**: WaitReady rejects empty SessionID without polling

```
WaitReady(SessionID="") -> error; StatusPollCalls==0
```

## Steps

1. Clear SessionID; still provide StatusFn so only id gate is tested.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "   "
	req.StatusPollHold = statusReadyFixture()
	return nil
}
```
