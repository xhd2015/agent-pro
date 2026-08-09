# Scenario

**Feature**: all sessions older than Recent window → empty list

```
Recent=30m; sessions at Now-2h and Now-3h
  -> empty + "No sessions found"
```

## Preconditions

- RecentSet, Recent=30m; all last_active outside window.
- WantFormat true.

## Steps

1. Write two old sessions.
2. Recent=30m, WantFormat=true.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.RecentSet = true
	req.Recent = 30 * time.Minute
	req.WantFormat = true
	writeListSession(t, req.GrokHome, idA1, atFixed(-2*time.Hour), cwdA, "old A")
	writeListSession(t, req.GrokHome, idB1, atFixed(-3*time.Hour), cwdB, "old B")
	return nil
}
```
