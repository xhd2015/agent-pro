# Scenario

**Feature**: snapshot change mid-window resets idle hits → no SoftExit

```
idle @0, idle @5s, changed snap @5s, idle @10s, idle @15s -> SoftExit=0
```

## Steps

1. Two idle hits, then a different resting snapshot (reset), then two more idles
   (not enough for SoftExit within this step list).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Steps = []TickStep{
		idleAt(0),
		idleAt(defaultFakeTimeout / 2),
		changedAt(defaultFakeTimeout/2, "chrome moved"),
		idleAt(defaultFakeTimeout),
		idleAt(defaultFakeTimeout + defaultFakeGrace),
	}
	return nil
}
```
