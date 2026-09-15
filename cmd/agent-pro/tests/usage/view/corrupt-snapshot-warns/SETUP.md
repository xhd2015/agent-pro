# Scenario

**Feature**: an unparseable snapshot file is skipped with a warning

```
usages/grok/.../10-00-00-snapshot.jsonl   one good line
usages/codex/.../01-02-03-snapshot.jsonl  "{not json}"
usage view --port 0
stderr: warning: usage view: skipping <path>: parse <path>: ...
GET /api/summary -> 200, grok still charted, errors[] names the file
```

## Preconditions

- A half-written line is what a crashed collect leaves behind; the dashboard has to
  survive it, since a blank dashboard is a worse answer than a partial one.
- The warnings are printed before the serving line, so by the time the harness has
  the URL the warning is already on stderr.

## Steps

1. Plant one readable grok snapshot and one unreadable codex file.
2. Probe `/api/summary`.

```go
import (
"os"
"path/filepath"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
WriteViewSnapshot(t, req, usage.Grok, time.Now().Add(-time.Hour),
map[string]float64{"used_percent": 42}, Display("42% used", "58% remaining"), 178, true, "")

today := time.Now().Local().Format("2006-01-02")
dir := filepath.Join(req.Home, "usages", string(usage.Codex), today)
if err := os.MkdirAll(dir, 0o755); err != nil {
t.Fatalf("mkdir codex day dir: %v", err)
}
broken := filepath.Join(dir, "01-02-03"+usage.SnapshotSuffix)
if err := os.WriteFile(broken, []byte("{not json}\n"), 0o644); err != nil {
t.Fatalf("write broken snapshot: %v", err)
}
req.HTTPPath = "/api/summary"
startViewServer(t, req)
return nil
}
```
