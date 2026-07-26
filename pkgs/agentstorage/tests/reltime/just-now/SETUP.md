# Scenario

**Feature**: ages under 1 second (and future clamped to zero) show `just now`

```
FormatRelativeAge(now, now-0.5s) -> "just now"
FormatRelativeAge(now, now)      -> "just now"
FormatRelativeAge(now, now+5s)  -> "just now"  # d < 0 treated as 0
```

## Preconditions

- Threshold is strict: `d < 1s` → `just now` (no ` ago` suffix).
- Negative age is clamped to zero before the threshold check.

## Steps

1. Three cases: half-second past, equal times, five seconds in the future.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	now := req.Now
	req.Cases = []FormatCase{
		{Name: "half_second", Target: ageTarget(now, 500*time.Millisecond), Want: "just now"},
		{Name: "exact_equal", Target: now, Want: "just now"},
		{Name: "future_clamped", Target: now.Add(5 * time.Second), Want: "just now"},
	}
	return nil
}
```
