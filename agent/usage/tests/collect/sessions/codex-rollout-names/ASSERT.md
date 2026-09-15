## Expected

- Only `rollout-*.jsonl` files count: the stray `notes.txt` is ignored, so
  `total` is 2.
- `oldest` is the earlier rollout's name decoded as local time
  (2026-09-10T09:00:00 local) — a codex rollout name is written in the machine's
  local zone, so the stored UTC instant depends on that zone.
- `newest` is the newest modification time, not the name: 2026-09-12T15:30:00
  local is later than the other file but the earlier file was still touched more
  recently, and activity is what the newest bound reports.
- Both bounds are set, so a consumer can plot an active span.

## Side Effects

- None.

## Errors

- None.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

codex := resp.Record(t, usage.Codex)
AssertSessions(t, codex, 2,
time.Date(2026, 9, 10, 9, 0, 0, 0, time.Local).UTC().Format(time.RFC3339),
req.Now.Add(-1*time.Hour).UTC().Format(time.RFC3339))
}
```
