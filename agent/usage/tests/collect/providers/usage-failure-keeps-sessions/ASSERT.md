## Expected

- All three records have `usage.ok: false` with a non-empty error, one per
  provider: grok names the fixture it could not read, codex names the last
  endpoint it tried (`codex-wham.json`, after `codex-usage.json` failed), and
  commandcode reports missing credentials.
- No record fabricates usage: `values` and `display` stay empty, so a consumer
  charting `values` sees a gap rather than a zero.
- The session blocks are still filled: two grok sessions (the group's baseline plus
  this leaf's), and one each for codex and the credential-less Command Code home,
  with the grok and Command Code timestamps coming from their transcripts. Counting
  only reads local files, so a fetch failure never costs us the session trend.
- The record order and the snapshot timestamp are unchanged.

## Side Effects

- None.

## Errors

- Recorded in the records, not returned: `Collect` itself never fails.

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

grok := resp.Record(t, usage.Grok)
AssertUsageFailed(t, grok, "fixture grok-billing.json")
codex := resp.Record(t, usage.Codex)
AssertUsageFailed(t, codex, "fixture codex-wham.json")
cc := resp.Record(t, usage.CommandCode)
AssertUsageFailed(t, cc, "not authenticated")

for _, rec := range resp.Records {
if len(rec.Usage.Values) != 0 {
t.Fatalf("%s recorded values for a failed fetch: %v", rec.Provider, rec.Usage.Values)
}
if len(rec.Usage.Display) != 0 {
t.Fatalf("%s recorded display strings for a failed fetch: %v", rec.Provider, rec.Usage.Display)
}
if !rec.TS.Equal(req.Now) {
t.Fatalf("%s ts = %s, want %s", rec.Provider, rec.TS, req.Now)
}
}

AssertSessions(t, grok, 2, "2026-09-10T09:00:00Z", "2026-09-11T10:00:00Z")
AssertSessions(t, cc, 1, "2026-09-11T09:00:00Z",
req.Now.Add(-24*time.Hour).UTC().Format(time.RFC3339))
if codex.Sessions.Total != 1 {
t.Fatalf("codex sessions total = %d, want 1", codex.Sessions.Total)
}
if codex.Sessions.Oldest == nil || codex.Sessions.Newest == nil {
t.Fatalf("codex sessions bounds = %v..%v, want both set", codex.Sessions.Oldest, codex.Sessions.Newest)
}
}
```
