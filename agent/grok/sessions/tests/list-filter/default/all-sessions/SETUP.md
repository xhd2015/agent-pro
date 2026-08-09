# Scenario

**Feature**: no filters returns every discovered session (high limit)

```
write sessions under cwdA, cwdB, cwdC
  -> ListWithOptions(Limit=10, no place/recent/active/grep)
  -> three sessions newest-first
```

## Preconditions

- Three sessions with distinct last_active timestamps.
- No PlaceCWDs, RecentSet, Active, or GrepSet.

## Steps

1. Write three sessions (newest idC1, middle idB1, oldest idA1).
2. Set Limit=10.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	writeListSession(t, req.GrokHome, idA1, atFixed(-3*time.Hour), cwdA, "session A old")
	writeListSession(t, req.GrokHome, idB1, atFixed(-1*time.Hour), cwdB, "session B mid")
	writeListSession(t, req.GrokHome, idC1, atFixed(-10*time.Minute), cwdC, "session C new")
	return nil
}
```
