# Scenario

**Feature**: missing idle-policy.json is found=false; watchdog never starts

```
ReadIdlePolicy(home, sess) -> found=false, no error
NewIdleWatchdog(false, …).Tick(idle past timeout+grace) -> SoftExit 0, Shutdown 0
```

## Steps

1. Do not write a policy file.
2. Read, then Tick idle samples at 0, 10s, 15s.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.WritePolicy = false
	req.TickAfterPolicy = true
	req.SessionID = "sess-policy-missing"
	req.Steps = []TickStep{
		idleAt(0),
		idleAt(defaultFakeTimeout),
		idleAt(defaultFakeTimeout + defaultFakeGrace),
	}
	return nil
}
```
