# Scenario

**Feature**: Active=true returns only sessions listed in active_sessions.json

```
two on-disk sessions; active list has only idA1
  -> ListWithOptions(Active=true) -> [idA1]
```

## Preconditions

- Both sessions discoverable; only one file-active.

## Steps

1. Write idA1 (older) and idB1 (newer).
2. writeActiveSessions(idA1) only.
3. Active=true, Limit=10.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.Active = true
	writeListSession(t, req.GrokHome, idA1, atFixed(-30*time.Minute), cwdA, "active one")
	writeListSession(t, req.GrokHome, idB1, atFixed(-10*time.Minute), cwdB, "inactive newer")
	writeActiveSessions(t, req.GrokHome, idA1)
	return nil
}
```
