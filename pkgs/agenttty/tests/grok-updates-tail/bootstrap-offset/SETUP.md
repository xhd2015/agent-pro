# Scenario

**Feature**: bootstrap read honors `startOffset` before WatchLine

```
startOffset=0  -> replay all pre-seeded bytes
startOffset=EOF -> skip bytes already on disk at tail start
```

## Preconditions

- Bootstrap semantics mirror `updatesTailStartOffset` fresh (0) vs stale (EOF) cases.
- No scheduled appends in these leaves unless noted.

## Steps

1. Leaf `Setup` writes pre-seeded content and sets `StartOffset` or `StartOffsetAtEOF`.
2. `Run` starts tail, holds briefly, cancels.
3. `Assert` checks which content was emitted.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.TailStartDelay = 100 * time.Millisecond
	req.HoldAfterSchedule = 200 * time.Millisecond
	return nil
}
```