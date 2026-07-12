# Scenario

**Feature**: a zero unit after the first non-zero stops the chain (Q1)

```
1h0m5s -> "1h ago"   # not "1h5s ago"
4d0h2m -> "4d ago"   # not "4d2m ago"
# omit the zero unit and everything smaller
```

## Preconditions

- Walk units largest→smallest (d, h, m, s).
- After first non-zero unit, if the next unit is 0, stop (do not skip zeros to find a later non-zero).

## Steps

1. Case 1h + 0m + 5s.
2. Case 4d + 0h + 2m.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	now := req.Now
	req.Cases = []FormatCase{
		{
			Name:   "hour_zero_min_five_sec",
			Target: ageTarget(now, 1*time.Hour+5*time.Second),
			Want:   "1h ago",
		},
		{
			Name:   "days_zero_hour_two_min",
			Target: ageTarget(now, 4*24*time.Hour+2*time.Minute),
			Want:   "4d ago",
		},
	}
	return nil
}
```
