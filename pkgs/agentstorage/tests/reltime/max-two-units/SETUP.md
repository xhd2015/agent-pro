# Scenario

**Feature**: at most two units are shown even when more are non-zero

```
4d5h12m -> "4d5h ago"  # minutes dropped (third unit)
```

## Preconditions

- After selecting two non-zero units, stop even if smaller units remain.
- Combined with zero-stop: only consecutive non-zeros from the first non-zero count.

## Steps

1. Case 4 days + 5 hours + 12 minutes.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	now := req.Now
	req.Cases = []FormatCase{
		{
			Name:   "four_days_five_hours_twelve_min",
			Target: ageTarget(now, 4*24*time.Hour+5*time.Hour+12*time.Minute),
			Want:   "4d5h ago",
		},
	}
	return nil
}
```
