# Scenario

**Feature**: Limit <= 0 defaults to 20 after filters

```
25 sessions under cwdA; PlaceCWDs=[A]; Limit=0
  -> exactly 20 newest
```

## Preconditions

- 25 survivors after place filter.
- Limit left at 0 (default).

## Steps

1. Write 25 sessions under cwdA with decreasing last_active.
2. PlaceCWDs=[A], Limit=0.

```go
import (
	"fmt"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 0
	req.PlaceCWDs = []string{absPath(t, cwdA)}
	for i := 0; i < 25; i++ {
		// i=0 newest (Now-1m), i=24 oldest
		id := fmt.Sprintf("019f283a-aaaa-7aaa-aaaa-aaaaaaaaa%03d", i)
		writeListSession(t, req.GrokHome, id, atFixed(-time.Duration(i+1)*time.Minute), cwdA, fmt.Sprintf("s%d", i))
	}
	return nil
}
```
