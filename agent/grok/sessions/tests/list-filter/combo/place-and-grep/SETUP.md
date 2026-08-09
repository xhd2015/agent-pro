# Scenario

**Feature**: place AND grep — out-of-place content matches are ignored

```
# idB1 under B has GREP_TOKEN in title but PlaceCWDs=[A]
# idA1 under A has GREP_TOKEN
  -> only idA1
```

## Preconditions

- GrepSet with literal token; same search family as ListWithGrep (title hit).
- PlaceCWDs restricts to cwdA.

## Steps

1. Write matching title sessions under A and B.
2. PlaceCWDs=[A], Grep=GREP_PLACE_TOKEN, GrepSet=true.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.PlaceCWDs = []string{absPath(t, cwdA)}
	req.GrepSet = true
	req.Grep = "GREP_PLACE_TOKEN"
	writeListSession(t, req.GrokHome, idA1, atFixed(-20*time.Minute), cwdA, "A has GREP_PLACE_TOKEN")
	// Newer and matching content but wrong place — must drop.
	writeListSession(t, req.GrokHome, idB1, atFixed(-5*time.Minute), cwdB, "B has GREP_PLACE_TOKEN")
	// In place but no token.
	writeListSession(t, req.GrokHome, idA2, atFixed(-10*time.Minute), cwdA, "A unrelated")
	return nil
}
```
