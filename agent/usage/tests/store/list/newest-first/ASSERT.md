## Expected

- `List(0)` returns all four records, newest first, mixing providers: a reader
  sees the latest state of every provider at the top without sorting the files
  themselves.
- Records stamped with the same instant are ordered by provider name (`codex`
  before `grok`), so the order is stable rather than dependent on directory
  iteration.
- Each record keeps the path it came from, so a caller can show or open the file.
- `List(2)` returns the two newest records only: truncation happens after the
  ordering, never before.

## Side Effects

- None.

## Errors

- None.

```go
import (
"fmt"
"strings"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

all := resp.Stored[0]
if got, want := sequence(all), "codex@2026-09-15T03:00:00Z, grok@2026-09-15T02:00:00Z, codex@2026-09-15T01:00:00Z, grok@2026-09-15T01:00:00Z"; got != want {
t.Fatalf("List(0) = %s, want %s", got, want)
}

for i, item := range all {
if i > 0 && item.Record.TS.After(all[i-1].Record.TS) {
t.Fatalf("record %d (%s) is newer than its predecessor (%s)", i, item.Record.TS, all[i-1].Record.TS)
}
if !strings.HasSuffix(item.Path, usage.SnapshotSuffix) {
t.Fatalf("record %d path = %q, want a snapshot file", i, item.Path)
}
if !strings.Contains(item.Path, string(item.Record.Provider)) {
t.Fatalf("record %d path = %q, want it under the %s provider directory", i, item.Path, item.Record.Provider)
}
}

truncated := resp.Stored[1]
if got, want := sequence(truncated), "codex@2026-09-15T03:00:00Z, grok@2026-09-15T02:00:00Z"; got != want {
t.Fatalf("List(2) = %s, want %s", got, want)
}
if got, want := truncated[1].Path, all[1].Path; got != want {
t.Fatalf("List(2) path = %s, want %s", got, want)
}
}

// sequence renders "provider@ts" for every record, for compact comparisons.
func sequence(records []usage.StoredRecord) string {
parts := make([]string, 0, len(records))
for _, item := range records {
parts = append(parts, fmt.Sprintf("%s@%s", item.Record.Provider, item.Record.TS.UTC().Format(time.RFC3339)))
}
return strings.Join(parts, ", ")
}
```
