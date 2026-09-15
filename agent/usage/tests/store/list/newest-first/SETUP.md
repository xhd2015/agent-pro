# Scenario

**Feature**: stored records come back newest first across providers and days

```
grok  @01:00 (1)   codex @01:00 (2)   grok @02:00 (3)   codex @03:00 (4)
List(0) -> codex@03:00, grok@02:00, codex@01:00, grok@01:00   # ties: provider name
List(2) -> the two newest only
```

## Preconditions

- Two providers and three instants, with one instant shared, so both the ordering
  and the tie-break are exercised.
- Each instant is a full hour apart, so every record lands in its own file.

## Steps

1. Append four records.
2. List everything, then list only the two newest.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
first := time.Date(2026, 9, 15, 1, 0, 0, 0, time.UTC)
second := time.Date(2026, 9, 15, 2, 0, 0, 0, time.UTC)
third := time.Date(2026, 9, 15, 3, 0, 0, 0, time.UTC)
req.Appends = []usage.Record{
Snapshot(usage.Grok, first, 1),
Snapshot(usage.Codex, first, 2),
Snapshot(usage.Grok, second, 3),
Snapshot(usage.Codex, third, 4),
}
req.ListLasts = []int{0, 2}
return nil
}
```
