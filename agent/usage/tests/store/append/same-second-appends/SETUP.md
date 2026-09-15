# Scenario

**Feature**: a second cycle inside the same second appends a line instead of overwriting the snapshot

```
Append(grok, 03:58:57Z, 1) -> usages/grok/<day>/03-58-57-snapshot.jsonl
Append(grok, 03:58:57Z, 2) -> same file, second line
Append(grok, 03:58:57Z, 3) -> same file, third line
Append(codex, 03:58:57Z, 9) -> its own provider file
```

## Preconditions

- All four records stamp the same UTC second, as a fast `--cron` loop would.

## Steps

1. Append three grok records and one codex record at the same instant.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
ts := time.Date(2026, 9, 15, 3, 58, 57, 0, time.UTC)
req.Appends = []usage.Record{
Snapshot(usage.Grok, ts, 1),
Snapshot(usage.Grok, ts, 2),
Snapshot(usage.Grok, ts, 3),
Snapshot(usage.Codex, ts, 9),
}
return nil
}
```
