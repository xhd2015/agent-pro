# Scenario

**Feature**: custom list limit caps result count

```
# five rollout sessions on disk
writeRolloutSession x5 -> sessions.List(codexHome, 3)

# only three newest sessions returned
[]Session (len=3)
```

## Preconditions

- Five rollout files exist with distinct timestamps.

## Steps

1. Create 5 sessions from `2026-06-22T10:00:00.000Z` to `2026-06-22T10:04:00.000Z`.
2. Set `req.Limit = 3`.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Limit = 3
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("01900002-0000-7000-8000-%012d", i+1)
		ts := fmt.Sprintf("2026-06-22T10:%02d:00.000Z", i)
		writeRolloutSession(t, req.CodexHome, id, ts, "/workspace/a")
	}
	return nil
}
```