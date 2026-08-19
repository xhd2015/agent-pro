# Scenario

**Feature**: Codex rollout jsonl size gate on top of chrome SampleIsIdle

```
chrome-idle samples + LogFound
  size grows between Ticks -> not idle (reset)
  size stable -> still SoftExit after 3 hits
```

## Steps

1. Grouping is under watchdog (`Op=watchdog`).
2. Leaves set `Steps` with `LogFound` / `LogBytes` on otherwise idle samples.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	return nil
}

func idleLogAt(at time.Duration, size int64) TickStep {
	s := defaultIdleSample()
	s.LogFound = true
	s.LogBytes = size
	return TickStep{At: at, Sample: s}
}
```
