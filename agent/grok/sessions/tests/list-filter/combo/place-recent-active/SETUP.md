# Scenario

**Feature**: place × recent × active three-way AND smoke

```
# only session under A, recent, and file-active survives
PlaceCWDs=[A] + Recent=1h + Active
```

## Preconditions

- Fixtures cover distractors for each dimension.

## Steps

1. Write:
   - idA1: A, recent, active → keep
   - idA2: A, recent, inactive
   - idA3: A, old, active
   - idB1: B, recent, active
2. Place A, Recent 1h, Active=true.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.PlaceCWDs = []string{absPath(t, cwdA)}
	req.RecentSet = true
	req.Recent = time.Hour
	req.Active = true
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "keep three-way")
	writeListSession(t, req.GrokHome, idA2, atFixed(-15*time.Minute), cwdA, "A recent inactive")
	writeListSession(t, req.GrokHome, idA3, atFixed(-3*time.Hour), cwdA, "A old active")
	writeListSession(t, req.GrokHome, idB1, atFixed(-5*time.Minute), cwdB, "B recent active")
	writeActiveSessions(t, req.GrokHome, idA1, idA3, idB1)
	return nil
}
```
