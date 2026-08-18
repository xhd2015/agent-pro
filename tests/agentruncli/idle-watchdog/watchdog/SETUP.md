# Scenario

**Feature**: armed idle watchdog Tick on a fake clock

```
NewIdleWatchdog(found=true, exit_on_idle=true, timeout=10s)
  Now/Sample injected -> Tick
  continuous idle -> SoftExit once, then Shutdown after grace 5s
  non-idle / reset / late first idle -> no premature exit
```

## Preconditions

- Armed policy: `NewIdleWatchdog(true, {ExitOnIdle:true, IdleTimeout:10s}, cfg)`.
- `WatchdogGrace` left 0 so product default 5s applies.
- No real time.Sleep.

## Steps

1. Grouping sets `Op=watchdog`.
2. Leaves set `Steps` (fake-clock samples).

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opWatchdog
	req.WatchdogTimeout = defaultFakeTimeout
	req.WatchdogGrace = 0
	return nil
}

func busySample() agentruncli.IdleSample {
	s := defaultIdleSample()
	s.Screen = "busy"
	return s
}

func occupiedSample() agentruncli.IdleSample {
	s := defaultIdleSample()
	s.InputBox = "occupied"
	return s
}
```
