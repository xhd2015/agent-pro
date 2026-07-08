# Scenario

**Bug**: rapid follow-ups within the 90s tail overlap window duplicate user events

```
single agentsync.EnsureGrokSync worker tails updates.jsonl
  -> turn 1 + turn 2 appends while worker alive
  -> one user message per distinct prompt text
```

## Preconditions

- Simulates web/CLI rapid follow-up timing (sub-second gap, within 90s overlap).

## Steps

1. Grouping leaves start worker on empty or partial `updates.jsonl`.
2. Schedule multiple turn appends before worker stops.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	if req.HoldAfterSchedule <= 0 {
		req.HoldAfterSchedule = 1200 * time.Millisecond
	}
	return nil
}
```