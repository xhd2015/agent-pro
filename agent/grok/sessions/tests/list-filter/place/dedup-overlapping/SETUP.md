# Scenario

**Feature**: duplicate PlaceCWDs entries do not duplicate sessions

```
PlaceCWDs=[abs(A), abs(A)] + one session under A
  -> single result
```

## Preconditions

- One session under cwdA.
- PlaceCWDs lists the same abs path twice.

## Steps

1. Write idA1 under cwdA.
2. PlaceCWDs = [abs(A), abs(A)].

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	a := absPath(t, cwdA)
	req.Limit = 10
	req.PlaceCWDs = []string{a, a}
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "once only")
	return nil
}
```
