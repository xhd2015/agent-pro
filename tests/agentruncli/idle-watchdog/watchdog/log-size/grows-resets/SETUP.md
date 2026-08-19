# Scenario

**Bug**: chrome-idle while Codex jsonl is still growing must not SoftExit

```
idle+log 100 @0, idle+log 200 @5s, idle+log 300 @10s -> SoftExit=0
```

Crime scene: seatalk-local-bot-c6f91a2f… turn 2 ran ~10m of exec while TTY
looked sendable/idle; rollout jsonl size was still increasing at abort.

## Steps

1. Three chrome-idle samples with increasing `LogBytes`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Steps = []TickStep{
		idleLogAt(0, 100),
		idleLogAt(defaultFakeTimeout/2, 200),
		idleLogAt(defaultFakeTimeout, 300),
	}
	return nil
}
```
