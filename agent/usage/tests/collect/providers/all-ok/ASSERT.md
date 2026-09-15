## Expected

- Three records, in the fixed provider order `grok`, `codex`, `commandcode`, so a
  stored snapshot set stays comparable run over run.
- Every record carries the injected clock, truncated to the second and in UTC.
- grok: billing endpoint, 73 of 100 monthly credits used (73% used, 27%
  remaining, resets 2026-09-01T00:00:00Z), display summary `73% used · monthly`.
- codex: `codex/usage` endpoint, 62% used with 38% remaining, plan `business`.
- Both grok and codex expose exactly the values listed below: an unknown field is
  omitted rather than written as -1 or 0.
- Sessions: one per provider. grok oldest/newest come from the summary timestamps,
  codex and commandcode newest from the transcript's modification time.

## Side Effects

- No files written: the store is not part of `Collect`.

## Errors

- None.

```go
import (
"reflect"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

if got, want := resp.Providers(), []string{"grok", "codex", "commandcode"}; !reflect.DeepEqual(got, want) {
t.Fatalf("providers = %v, want %v", got, want)
}

for _, rec := range resp.Records {
if !rec.TS.Equal(req.Now) {
t.Fatalf("%s ts = %s, want %s", rec.Provider, rec.TS, req.Now)
}
if rec.Schema != usage.SchemaID {
t.Fatalf("%s schema = %q, want %q", rec.Provider, rec.Schema, usage.SchemaID)
}
}

grok := resp.Record(t, usage.Grok)
AssertUsageOK(t, grok)
if grok.Usage.Endpoint != "billing" {
t.Fatalf("grok endpoint = %q, want billing", grok.Usage.Endpoint)
}
AssertValues(t, grok, map[string]float64{
"used":              73,
"monthly_limit":     100,
"used_percent":      73,
"remaining_percent": 27,
"period_start":      float64(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC).Unix()),
"period_end":        float64(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Unix()),
"reset_at":          float64(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Unix()),
})
if got, want := Display(t, grok, "usage"), "73% used · monthly"; got != want {
t.Fatalf("grok usage headline = %q, want %q", got, want)
}
if got, want := Display(t, grok, "period"), "monthly"; got != want {
t.Fatalf("grok period = %q, want %q", got, want)
}
if got, want := Display(t, grok, "email"), "grok@example.com"; got != want {
t.Fatalf("grok email = %q, want %q", got, want)
}
AssertSessions(t, grok, 1, "2026-09-10T09:00:00Z", "2026-09-10T10:00:00Z")

codex := resp.Record(t, usage.Codex)
AssertUsageOK(t, codex)
if codex.Usage.Endpoint != "codex/usage" {
t.Fatalf("codex endpoint = %q, want codex/usage", codex.Usage.Endpoint)
}
AssertValues(t, codex, map[string]float64{
"used_percent":      62,
"remaining_percent": 38,
"reset_at":          float64(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Unix()),
})
if got, want := Display(t, codex, "usage"), "62% used"; got != want {
t.Fatalf("codex usage headline = %q, want %q", got, want)
}
if got, want := Display(t, codex, "plan_type"), "business"; got != want {
t.Fatalf("codex plan type = %q, want %q", got, want)
}
// A codex rollout file name carries the session start in local time.
codexStarted := time.Date(2026, 9, 10, 9, 0, 0, 0, time.Local)
AssertSessions(t, codex, 1,
codexStarted.UTC().Format(time.RFC3339),
req.Now.Add(-24*time.Hour).UTC().Format(time.RFC3339))

cc := resp.Record(t, usage.CommandCode)
AssertUsageOK(t, cc)
if got, want := cc.Usage.Endpoint, usage.EndpointCommandCodeUsage; got != want {
t.Fatalf("commandcode endpoint = %q, want %q", got, want)
}
AssertSessions(t, cc, 1, "2026-09-10T09:00:00Z",
req.Now.Add(-24*time.Hour).UTC().Format(time.RFC3339))
}
```
