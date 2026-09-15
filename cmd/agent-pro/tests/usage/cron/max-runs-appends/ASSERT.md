## Expected

- Exit 0 after the third cycle: `--max-runs` ends the loop as a success, so a
  bounded run is testable and scriptable.
- stdout holds nine JSON records, nine = three cycles × three providers, in cycle
  order: the loop cycles immediately, then per schedule.
- Each provider has exactly three stored records. They may share one snapshot file
  when they land in the same second, which is why the count is per provider and
  not per file.
- The stored timestamps for one provider are non-decreasing, so a reader can order
  cycles without guessing.
- stderr has at least one `usage collect: next run at <UTC>` line, so a cron job's
  log shows when the next cycle is due.

## Side Effects

- Creates `$AGENT_PRO_HOME/usages/<provider>/<today>/` with one or more files per
  provider.

## Errors

- None: every provider answered in every cycle.

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

printed := DecodeJSONLines(t, resp.Stdout)
if len(printed) != 9 {
t.Fatalf("printed %d records, want 9 (3 cycles x 3 providers)", len(printed))
}
for i, record := range printed {
if want := usage.AllProviders()[i%3]; record.Provider != want {
t.Fatalf("record %d provider = %s, want %s (cycle order)", i, record.Provider, want)
}
if !record.Usage.OK {
t.Fatalf("record %d (%s) usage not ok: %s", i, record.Provider, record.Usage.Error)
}
}

AssertStderrContains(t, resp, "usage collect: next run at ")

for _, provider := range usage.AllProviders() {
stored := resp.RecordsFor(t, provider)
if len(stored) != 3 {
t.Fatalf("%s stored %d records, want 3", provider, len(stored))
}
for i := 1; i < len(stored); i++ {
if stored[i].TS.Before(stored[i-1].TS) {
t.Fatalf("%s cycle %d ts = %s, older than cycle %d (%s)",
provider, i, stored[i].TS, i-1, stored[i-1].TS)
}
}
if last := stored[len(stored)-1].TS; time.Since(last) > 2*time.Minute {
t.Fatalf("%s last cycle ts = %s, want within 2m of now", provider, last)
}
}
}
```
