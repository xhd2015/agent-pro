# Scenario

**Feature**: ages that collapse to a single non-zero unit

```
1s  -> "1s ago"
2s  -> "2s ago"
1h  -> "1h ago"
90d -> "90d ago"
# short labels only: s m h d
```

## Preconditions

- When only one unit is non-zero after flooring, output is that unit plus ` ago`.
- Large day counts are allowed (`90d`), not capped to weeks/months.

## Steps

1. Cases for 1s, 2s, 1h, and 90 days exactly.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	now := req.Now
	req.Cases = []FormatCase{
		{Name: "one_second", Target: ageTarget(now, 1*time.Second), Want: "1s ago"},
		{Name: "two_seconds", Target: ageTarget(now, 2*time.Second), Want: "2s ago"},
		{Name: "one_hour", Target: ageTarget(now, 1*time.Hour), Want: "1h ago"},
		{Name: "ninety_days", Target: ageTarget(now, 90*24*time.Hour), Want: "90d ago"},
	}
	return nil
}
```
