# Scenario

**Feature**: armed idle.Watchdog Tick on a fake clock

```
idle.New(found=true, exit_on_idle=true, timeout=10s)
  Now/Snapshot/ProbeOccupied injected -> Tick
  stable+empty x3 -> SoftExit once, then Shutdown after grace 5s
  changed / occupied mid-window -> reset, no premature exit
```

## Preconditions

- Armed policy: `idle.New(true, {ExitOnIdle:true, IdleTimeout:10s}, cfg)`.
- `WatchdogGrace` left 0 so product default 5s applies.
- No real time.Sleep.

## Steps

1. Grouping sets `Op=watchdog`.
2. Leaves set `Steps` (fake-clock samples).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opWatchdog
	req.WatchdogTimeout = defaultFakeTimeout
	req.WatchdogGrace = 0
	return nil
}
```
