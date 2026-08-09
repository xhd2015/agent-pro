# Scenario

**Feature**: recent AND active

```
# recent+active kept
# recent inactive dropped
# old active dropped
Recent=1h + Active
```

## Preconditions

- Three sessions: recent-active, recent-inactive, old-active.

## Steps

1. Write idA1 recent active; idB1 recent inactive; idC1 old active.
2. Active list = idA1, idC1; Recent=1h.

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
	req.Active = true
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "recent active")
	writeListSession(t, req.GrokHome, idB1, atFixed(-15*time.Minute), cwdB, "recent inactive")
	writeListSession(t, req.GrokHome, idC1, atFixed(-3*time.Hour), cwdC, "old active")
	writeActiveSessions(t, req.GrokHome, idA1, idC1)
	return nil
}
```
