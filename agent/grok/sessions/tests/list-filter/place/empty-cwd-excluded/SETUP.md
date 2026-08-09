# Scenario

**Feature**: session with empty info.cwd never matches place filter

```
session with info.cwd="" + PlaceCWDs=[abs(cwdA)]
  + normal session under cwdA
  -> only normal session
```

## Preconditions

- Empty-cwd session still has on-disk layout under a storage key.
- Place filter is ON.

## Steps

1. Write empty-cwd session (newer) and normal cwdA session.
2. PlaceCWDs = abs(cwdA).

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.PlaceCWDs = []string{absPath(t, cwdA)}
	// Empty cwd session is newer but must not match place.
	writeListSessionOpts(t, req.GrokHome, idE1, atFixed(-5*time.Minute), "", "empty cwd", listSessionOpts{
		CWDEmpty:   true,
		StorageCWD: "/tmp/list-filter-empty-cwd-storage",
	})
	writeListSession(t, req.GrokHome, idA1, atFixed(-15*time.Minute), cwdA, "normal A")
	return nil
}
```
