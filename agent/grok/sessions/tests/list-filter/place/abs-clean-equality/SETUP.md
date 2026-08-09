# Scenario

**Feature**: place match uses Abs+Clean on both sides (trailing slash OK)

```
session info.cwd = abs(cwdA)
PlaceCWDs = [abs(cwdA) + "/"]  # trailing slash
  -> still matches after Clean
```

## Preconditions

- Session stored with cleaned absolute cwd (writeListSession Abs).
- PlaceCWD intentionally has a trailing separator after abs path.

## Steps

1. Write session under cwdA.
2. Set PlaceCWDs to absPath(cwdA) + string(os.PathSeparator) (or "/").
3. Limit=10.

```go
import (
	"os"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	// Trailing slash must still match Abs+Clean(session.CWD).
	req.PlaceCWDs = []string{absPath(t, cwdA) + string(os.PathSeparator)}
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "clean match")
	writeListSession(t, req.GrokHome, idB1, atFixed(-5*time.Minute), cwdB, "other")
	return nil
}
```
