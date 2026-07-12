# Scenario

**Feature**: two consecutive non-zero units are both shown

```
65s   -> "1m5s ago"
1h2m  -> "1h2m ago"
# no spaces between units; space only before "ago"
```

## Preconditions

- Max two units; both non-zero so zero-stop does not apply.
- Units joined without interstitial spaces.

## Steps

1. Case 65 seconds → minutes+seconds.
2. Case 1 hour 2 minutes → hours+minutes (exact).

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	now := req.Now
	req.Cases = []FormatCase{
		{Name: "sixty_five_seconds", Target: ageTarget(now, 65*time.Second), Want: "1m5s ago"},
		{Name: "one_hour_two_min", Target: ageTarget(now, 1*time.Hour+2*time.Minute), Want: "1h2m ago"},
	}
	return nil
}
```
