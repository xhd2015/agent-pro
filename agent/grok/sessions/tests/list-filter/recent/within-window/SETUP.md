# Scenario

**Feature**: Recent window keeps in-window last_active and drops older

```
Now=fixedNow, Recent=1h
  session Now-10m kept; Now-3h dropped
```

## Preconditions

- RecentSet=true, Recent=1h.
- Two sessions under same cwd (place not set).

## Steps

1. Write recent idA1 and old idB1.
2. Set Recent=time.Hour, RecentSet=true, Limit=10.

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
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "recent")
	writeListSession(t, req.GrokHome, idB1, atFixed(-3*time.Hour), cwdA, "too old")
	return nil
}
```
