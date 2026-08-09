# Scenario

**Feature**: recent AND grep — old matching content dropped

```
# old session title has token but last_active outside window
# recent session has token
  -> only recent
```

## Preconditions

- Recent=1h; Grep token present in both titles.

## Steps

1. Write old matching idB1 and recent matching idA1; recent non-match idA2.
2. Recent=1h, Grep=GREP_RECENT_TOKEN.

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
	req.GrepSet = true
	req.Grep = "GREP_RECENT_TOKEN"
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "fresh GREP_RECENT_TOKEN")
	writeListSession(t, req.GrokHome, idB1, atFixed(-3*time.Hour), cwdA, "stale GREP_RECENT_TOKEN")
	writeListSession(t, req.GrokHome, idA2, atFixed(-15*time.Minute), cwdA, "fresh no token")
	return nil
}
```
