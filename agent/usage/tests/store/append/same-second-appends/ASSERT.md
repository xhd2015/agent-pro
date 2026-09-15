## Expected

- The three grok appends return the same path: same provider, same second.
- That file holds three lines in append order, with the values 1, 2, 3. Nothing is
  overwritten, so a fast loop loses no cycle.
- The codex record goes to a different file: providers never share a snapshot file.

## Side Effects

- Creates the grok and codex day directories.

## Errors

- None.

```go
import (
"testing"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

for i := 0; i < 3; i++ {
if resp.Paths[i] != resp.Paths[0] {
t.Fatalf("path[%d] = %s, want %s (same second, same provider)", i, resp.Paths[i], resp.Paths[0])
}
}
if resp.Paths[3] == resp.Paths[0] {
t.Fatalf("grok and codex wrote to the same file %s", resp.Paths[0])
}

lines := resp.Lines[resp.Paths[0]]
if len(lines) != 3 {
t.Fatalf("grok file holds %d lines, want 3", len(lines))
}
for i, line := range lines {
record := DecodeLine(t, line)
if record.Provider != string(usage.Grok) {
t.Fatalf("line %d provider = %q, want grok", i, record.Provider)
}
if got, want := record.Usage.Values["used_percent"], float64(i+1); got != want {
t.Fatalf("line %d used_percent = %v, want %v (append order)", i, got, want)
}
if record.TS != req.Appends[i].TS.UTC().Format("2006-01-02T15:04:05Z07:00") {
t.Fatalf("line %d ts = %q, want the shared second", i, record.TS)
}
}

codexLines := resp.Lines[resp.Paths[3]]
if len(codexLines) != 1 {
t.Fatalf("codex file holds %d lines, want 1", len(codexLines))
}
if got, want := DecodeLine(t, codexLines[0]).Usage.Values["used_percent"], float64(9); got != want {
t.Fatalf("codex used_percent = %v, want %v", got, want)
}
}
```
