# Scenario

**Feature**: one append per provider cycle, at a path that encodes provider and local time

```
Append(Snapshot(grok, 2026-09-15T03:58:57Z, 73))  -> usages/grok/<local day>/<local time>-snapshot.jsonl
Append(Snapshot(codex, 2026-09-15T04:00:00Z, 62)) -> usages/codex/...
List(0)                                           -> both records, newest first
```

## Preconditions

- The root does not exist yet; `Append` creates it.
- `ts` is truncated to the second, so the file name and the record agree.

## Steps

1. Append one grok and one codex record.
2. Read both files back as raw lines and list the store.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Appends = []usage.Record{
Snapshot(usage.Grok, time.Date(2026, 9, 15, 3, 58, 57, 0, time.UTC), 73),
Snapshot(usage.Codex, time.Date(2026, 9, 15, 4, 0, 0, 0, time.UTC), 62),
}
req.ListLasts = []int{0}
return nil
}
```
