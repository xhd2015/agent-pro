## Expected

- Exit 0 and exactly two JSON lines, no table and no trailer: stdout stays
  machine-readable.
- The lines are the records as stored, newest first, with the fields a consumer
  reads: schema, `ts` in UTC, provider, the usage block and the session block.
- The values survive the store round trip: the usage headline and the session total
  read back are the ones that were planted.

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
AssertExitCode(t, resp, 0)

records := DecodeJSONLines(t, resp.Stdout)
if len(records) != 2 {
t.Fatalf("got %d records, want 2:\n%s", len(records), resp.Stdout)
}

newest := records[0]
if newest.Provider != usage.Codex {
t.Fatalf("first record provider = %s, want codex (newest first)", newest.Provider)
}
if want := time.Date(2026, 9, 15, 2, 0, 0, 0, time.UTC); !newest.TS.Equal(want) {
t.Fatalf("first record ts = %s, want %s", newest.TS, want)
}
if got, want := newest.Usage.Display["usage"], "20% used"; got != want {
t.Fatalf("codex usage headline = %q, want %q", got, want)
}
if got, want := newest.Sessions.Total, 2; got != want {
t.Fatalf("codex sessions total = %d, want %d", got, want)
}
if newest.Schema != usage.SchemaID {
t.Fatalf("codex schema = %q, want %q", newest.Schema, usage.SchemaID)
}

oldest := records[1]
if oldest.Provider != usage.Grok {
t.Fatalf("second record provider = %s, want grok", oldest.Provider)
}
if got, want := oldest.Sessions.Total, 1; got != want {
t.Fatalf("grok sessions total = %d, want %d", got, want)
}
}
```
