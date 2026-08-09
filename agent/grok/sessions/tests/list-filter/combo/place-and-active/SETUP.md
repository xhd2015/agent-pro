# Scenario

**Feature**: place AND active

```
# in-place active kept
# in-place inactive dropped
# other-cwd active dropped
PlaceCWDs=[A] + Active
```

## Preconditions

- idA1 in A marked active; idA2 in A inactive; idB1 in B marked active.

## Steps

1. Write three sessions; active list = idA1 and idB1.
2. PlaceCWDs=[A], Active=true.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.PlaceCWDs = []string{absPath(t, cwdA)}
	req.Active = true
	writeListSession(t, req.GrokHome, idA1, atFixed(-20*time.Minute), cwdA, "A active keep")
	writeListSession(t, req.GrokHome, idA2, atFixed(-10*time.Minute), cwdA, "A inactive drop")
	writeListSession(t, req.GrokHome, idB1, atFixed(-5*time.Minute), cwdB, "B active drop")
	writeActiveSessions(t, req.GrokHome, idA1, idB1)
	return nil
}
```
