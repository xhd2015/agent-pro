# Scenario

**Feature**: last_active exactly at Now−Recent is kept (inclusive bound)

```
Recent=1h; last_active = fixedNow - 1h exactly
  -> session kept
```

## Preconditions

- Cutoff = Now - Recent; comparison is `!last_active.Before(cutoff)`.
- Companion session one second older is dropped.

## Steps

1. Write boundary idA1 at exactly -1h and older idB1 at -1h-1s.
2. Recent=1h, RecentSet=true.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.RecentSet = true
	req.Recent = time.Hour
	// Exact boundary — must be kept.
	writeListSession(t, req.GrokHome, idA1, atFixed(-1*time.Hour), cwdA, "on boundary")
	// Just outside — must drop.
	writeListSession(t, req.GrokHome, idB1, atFixed(-1*time.Hour-time.Second), cwdA, "just outside")
	return nil
}
```
