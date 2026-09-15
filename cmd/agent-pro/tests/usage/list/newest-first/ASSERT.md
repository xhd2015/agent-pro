## Expected

- Exit 0 and one header row plus exactly two data rows: `--last` bounds the output.
- The two rows are the newest snapshots, newest first (`grok` at 03:00 before
  `codex` at 02:00): the pair comes from the newest instants, not from whichever
  provider directory sorts first.
- Each row is `ts | provider | usage | sessions | file` in whitespace-separated
  columns: the timestamp is RFC3339 UTC, the usage column is the stored headline,
  the sessions column is the stored total, and the file column is the snapshot file
  that holds the record.
- The file column points at the same file that was listed, so a reader can open it.

## Side Effects

- None: listing never writes.

## Errors

- None.

```go
import (
"strings"
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertExitCode(t, resp, 0)

lines := resp.StdoutLines()
header := strings.Fields(lines[0])
for _, column := range []string{"ts", "provider", "usage", "sessions", "file"} {
if !contains(header, column) {
t.Fatalf("header %v has no %q column", header, column)
}
}

if len(lines) != 3 {
t.Fatalf("got %d lines, want a header plus 2 rows:\n%s", len(lines), resp.Stdout)
}

newest := strings.Fields(lines[1])
if want := []string{"2026-09-15T03:00:00Z", "grok", "30%", "used", "3"}; !hasPrefix(newest, want) {
t.Fatalf("row 1 = %v, want it to start with %v", newest, want)
}
if got := newest[len(newest)-1]; !strings.HasSuffix(got, "/usages/grok/2026-09-15/11-00-00-snapshot.jsonl") {
t.Fatalf("row 1 file column = %q, want the 11-00-00 snapshot of grok", got)
}

second := strings.Fields(lines[2])
if want := []string{"2026-09-15T02:00:00Z", "codex", "20%", "used", "2"}; !hasPrefix(second, want) {
t.Fatalf("row 2 = %v, want it to start with %v", second, want)
}
if got := second[len(second)-1]; !strings.HasSuffix(got, "/usages/codex/2026-09-15/10-00-00-snapshot.jsonl") {
t.Fatalf("row 2 file column = %q, want the codex 10-00-00 snapshot", got)
}

if strings.Contains(resp.Stdout, "09-00-00-snapshot.jsonl") {
t.Fatalf("--last 2 included the oldest snapshot:\n%s", resp.Stdout)
}
}

func contains(values []string, want string) bool {
for _, value := range values {
if value == want {
return true
}
}
return false
}

func hasPrefix(fields []string, prefix []string) bool {
if len(fields) < len(prefix) {
return false
}
for i, want := range prefix {
if fields[i] != want {
return false
}
}
return true
}
```
