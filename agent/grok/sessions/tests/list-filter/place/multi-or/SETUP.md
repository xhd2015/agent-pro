# Scenario

**Feature**: multiple PlaceCWDs form an OR set

```
sessions A, B, C + PlaceCWDs=[A,B]
  -> A and B kept; C excluded
```

## Preconditions

- Three sessions under cwdA, cwdB, cwdC.
- PlaceCWDs lists abs(cwdA) and abs(cwdB) only.

## Steps

1. Write one session per cwd with distinct last_active.
2. Set PlaceCWDs to [A,B], Limit=10.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.PlaceCWDs = []string{absPath(t, cwdA), absPath(t, cwdB)}
	writeListSession(t, req.GrokHome, idA1, atFixed(-30*time.Minute), cwdA, "A")
	writeListSession(t, req.GrokHome, idB1, atFixed(-20*time.Minute), cwdB, "B")
	writeListSession(t, req.GrokHome, idC1, atFixed(-10*time.Minute), cwdC, "C excluded")
	return nil
}
```
