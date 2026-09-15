## Expected

- Each append returns a path of the shape
  `<root>/<provider>/<local date>/<local time>-snapshot.jsonl`, so the tree can be
  browsed per provider and per day with `ls`.
- The date and time directories are the record's timestamp in local time; the
  `ts` inside the line stays UTC, so a snapshot written here and one written on
  another machine compare directly.
- The stored `ts` is an RFC3339 UTC string with no fractional seconds.
- Exactly one line per file, and each line is one complete record: schema,
  provider, usage values and display strings, and the session block.
- `List(0)` returns both records, newest first, with the same values that went in
  (a store round trip loses nothing).

## Side Effects

- Creates `<root>/<provider>/<local date>/` directories.

## Errors

- None.

```go
import (
"regexp"
"strings"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

if len(resp.Paths) != 2 {
t.Fatalf("paths = %v, want 2", resp.Paths)
}
for i, path := range resp.Paths {
if want := LayoutPath(req.Root, req.Appends[i]); path != want {
t.Fatalf("path[%d] = %s, want %s", i, path, want)
}
if !strings.HasSuffix(path, usage.SnapshotSuffix) {
t.Fatalf("path %s does not end with %s", path, usage.SnapshotSuffix)
}
if lines := resp.Lines[path]; len(lines) != 1 {
t.Fatalf("%s holds %d lines, want 1", path, len(lines))
}
}

wantTimeName := regexp.MustCompile(`^\d{2}-\d{2}-\d{2}-snapshot\.jsonl$`)
for i, path := range resp.Paths {
base := path[strings.LastIndex(path, "/")+1:]
if !wantTimeName.MatchString(base) {
t.Fatalf("file name %q is not HH-MM-SS-snapshot.jsonl", base)
}
wantTS := req.Appends[i].TS.UTC().Format(time.RFC3339)
record := DecodeLine(t, resp.Lines[path][0])
if record.Schema != usage.SchemaID {
t.Fatalf("schema = %q, want %q", record.Schema, usage.SchemaID)
}
if record.TS != wantTS {
t.Fatalf("stored ts = %q, want %q", record.TS, wantTS)
}
if strings.Contains(record.TS, ".") {
t.Fatalf("stored ts %q carries sub-second precision", record.TS)
}
if record.Provider != string(req.Appends[i].Provider) {
t.Fatalf("stored provider = %q, want %q", record.Provider, req.Appends[i].Provider)
}
if !record.Usage.OK || record.Usage.Endpoint != "billing" {
t.Fatalf("stored usage block = %+v, want ok with the billing endpoint", record.Usage)
}
if got, want := record.Usage.Values["used_percent"], req.Appends[i].Usage.Values["used_percent"]; got != want {
t.Fatalf("stored used_percent = %v, want %v", got, want)
}
if got, want := record.Usage.Display["usage"], req.Appends[i].Usage.Display["usage"]; got != want {
t.Fatalf("stored usage display = %q, want %q", got, want)
}
if got, want := record.Sessions.Total, req.Appends[i].Sessions.Total; got != want {
t.Fatalf("stored sessions total = %d, want %d", got, want)
}
}

stored := resp.Stored[0]
if len(stored) != 2 {
t.Fatalf("List(0) returned %d records, want 2", len(stored))
}
if stored[0].Record.Provider != usage.Codex || stored[1].Record.Provider != usage.Grok {
t.Fatalf("List order = %s, %s; want codex then grok (newest first)",
stored[0].Record.Provider, stored[1].Record.Provider)
}
if got, want := stored[1].Record.Usage.Values["used_percent"], float64(73); got != want {
t.Fatalf("grok used_percent read back = %v, want %v", got, want)
}
if got, want := stored[1].Path, resp.Paths[0]; got != want {
t.Fatalf("read-back path = %s, want %s", got, want)
}
}
```
