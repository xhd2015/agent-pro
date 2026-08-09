# Scenario

**Feature**: place AND recent — both must pass

```
# in-place recent kept
# in-place old dropped
# other-cwd recent dropped
PlaceCWDs=[A] + Recent=1h
```

## Preconditions

- Three sessions covering the AND matrix cells needed.

## Steps

1. Write idA1 in A recent; idA2 in A old; idB1 in B recent.
2. PlaceCWDs=[A], Recent=1h.

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
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "A recent keep")
	writeListSession(t, req.GrokHome, idA2, atFixed(-3*time.Hour), cwdA, "A old drop")
	writeListSession(t, req.GrokHome, idB1, atFixed(-5*time.Minute), cwdB, "B recent drop")
	return nil
}
```
