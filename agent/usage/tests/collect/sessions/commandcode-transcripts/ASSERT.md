## Expected

- `total` is 2: `<uuid>.jsonl` transcripts count, and neither the
  `<uuid>.checkpoints.jsonl` sidecar (a different file name for the same session)
  nor `notes.md` does.
- `oldest` is 2026-09-11T09:00:00Z, read from the `createdAt` of a first record
  that has no `timestamp`, so both transcript shapes give a start time.
- `newest` is the latest modification time (2026-09-12, i.e. the injected clock
  minus two hours), which is when the session was last written.

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

cc := resp.Record(t, usage.CommandCode)
AssertSessions(t, cc, 2, "2026-09-11T09:00:00Z",
req.Now.Add(-2*time.Hour).UTC().Format(time.RFC3339))
}
```
