# Scenario

**Feature**: snapshot files are appended under a browsable provider/day tree

```
NewStore(<tmp>/usages)
Append(record) -> usages/<provider>/<local YYYY-MM-DD>/<local HH-MM-SS>-snapshot.jsonl
List(last)     -> newest-first records
```

## Preconditions

- The root does not exist before the first append: the store creates it.
- Nothing here reads `~/.agent-pro/usages`; every leaf uses a temp root.

## Steps

1. Root `Setup` points `req.Root` at a fresh temp path and leaves the leaf to
   choose the records.
2. `Run` appends `req.Appends` in order, reads every written file back as raw
   lines, and lists when the leaf asks.
3. Leaf `Setup` builds records with `Snapshot(provider, ts, percent)`.
4. Leaf `Assert` compares paths, raw lines, and records read back.

## Context

- Paths use local time (so `ls` groups a day's snapshots); `ts` inside a record is
  UTC (so records from different machines compare).
- Timestamps are truncated to the second everywhere: two runs in one second share
  a file name, and a record never carries sub-second precision.

```go
import (
"path/filepath"
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Root = filepath.Join(t.TempDir(), "usages")
return nil
}
```
