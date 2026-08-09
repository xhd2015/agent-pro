# Scenario

**Feature**: place filter with no matching cwd yields empty list

```
sessions under cwdA + PlaceCWDs=[abs(cwdB)]
  -> empty + FormatListTable "No sessions found"
```

## Preconditions

- At least one real session exists under a non-matching cwd.
- WantFormat true.

## Steps

1. Write session under cwdA.
2. PlaceCWDs = abs(cwdB) only.
3. WantFormat=true.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.PlaceCWDs = []string{absPath(t, cwdB)}
	req.WantFormat = true
	writeListSession(t, req.GrokHome, idA1, atFixed(-10*time.Minute), cwdA, "not in place")
	return nil
}
```
