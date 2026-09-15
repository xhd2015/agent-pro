## Expected

- 200 with the dashboard schema, the store root, and the server's default range.
- Three providers, sorted by name, each with its newest snapshot: grok 42% over two
  snapshots, codex 61%, commandcode 29% through the `usage_percent` key. A card
  reads the newest instant, so an older file can never win.
- Each provider entry carries the age in seconds from the record's own timestamp and
  the path of the file it came from, so the dashboard links back to the snapshot.
- `recent` lists all four records newest first across providers, not grouped by
  provider.
- `metrics` starts with the two derived metrics and includes `primary_percent`,
  which covers all three providers even though their keys differ; the raw
  `used_percent` covers only grok and codex.
- Session totals come from the records: 178, 92, 41.

## Side Effects

- None: the store is read.

## Errors

- None.

```go
import (
"strings"
"testing"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertHTTPStatus(t, resp, 200)
AssertNoWarnings(t, resp)

summary := Summary(t, resp)
if summary.Schema != "agent-pro/usage-view/v1" {
t.Fatalf("schema = %q, want agent-pro/usage-view/v1", summary.Schema)
}
if summary.SnapshotCount != 4 || len(summary.Recent) != 4 {
t.Fatalf("snapshot_count = %d recent = %d, want 4 and 4", summary.SnapshotCount, len(summary.Recent))
}
if summary.Root != req.Home+"/usages" {
t.Fatalf("root = %q, want %q", summary.Root, req.Home+"/usages")
}
if summary.Range != "7d" {
t.Fatalf("range = %q, want the server default 7d", summary.Range)
}

got := []string{}
for _, entry := range summary.Providers {
got = append(got, entry.Provider)
}
if want := "codex,commandcode,grok"; strings.Join(got, ",") != want {
t.Fatalf("providers = %v, want %s", got, want)
}

grok := Provider(t, summary, "grok")
if value := grok.Latest.Usage.Values["used_percent"]; value != 42 {
t.Fatalf("grok latest used_percent = %v, want the newest 42", value)
}
if grok.Count != 2 {
t.Fatalf("grok snapshot count = %d, want 2", grok.Count)
}
if grok.Latest.Sessions.Total != 178 {
t.Fatalf("grok sessions total = %d, want 178", grok.Latest.Sessions.Total)
}
if grok.AgeSeconds < 3500 || grok.AgeSeconds > 3700 {
t.Fatalf("grok age = %v, want about one hour", grok.AgeSeconds)
}
if !strings.Contains(grok.Latest.Path, "/usages/grok/") || !strings.HasSuffix(grok.Latest.Path, "-snapshot.jsonl") {
t.Fatalf("grok latest path = %q, want its snapshot file", grok.Latest.Path)
}

codex := Provider(t, summary, "codex")
if codex.Latest.Usage.Values["used_percent"] != 61 || codex.Latest.Sessions.Total != 92 {
t.Fatalf("codex = %+v, want 61%% over 92 sessions", codex.Latest)
}

cc := Provider(t, summary, "commandcode")
if _, ok := cc.Latest.Usage.Values["usage_percent"]; !ok {
t.Fatalf("commandcode values = %v, want the provider's own usage_percent key", cc.Latest.Usage.Values)
}
if cc.Latest.Sessions.Total != 41 {
t.Fatalf("commandcode sessions total = %d, want 41", cc.Latest.Sessions.Total)
}

order := []string{}
for _, record := range summary.Recent {
order = append(order, record.Provider)
}
if want := "commandcode,grok,codex,grok"; strings.Join(order, ",") != want {
t.Fatalf("recent order = %v, want newest first %s", order, want)
}
if !summary.Recent[0].TS.After(summary.Recent[3].TS) {
t.Fatalf("recent is not newest first: %s then %s", summary.Recent[0].TS, summary.Recent[3].TS)
}

if len(summary.Metrics) == 0 {
t.Fatal("metrics is empty, so the dashboard menu would have nothing to offer")
}
if summary.Metrics[0].Name != "primary_percent" || summary.Metrics[1].Name != "sessions_total" {
t.Fatalf("metrics start with %s, %s, want the derived pair first",
summary.Metrics[0].Name, summary.Metrics[1].Name)
}
for _, metric := range summary.Metrics {
switch metric.Name {
case "primary_percent":
if len(metric.Providers) != 3 || !metric.Chartable || metric.Unit != "%" {
t.Fatalf("primary_percent = %+v, want three providers in percent", metric)
}
case "used_percent":
if want := "codex,grok"; strings.Join(metric.Providers, ",") != want {
t.Fatalf("used_percent providers = %v, want %s", metric.Providers, want)
}
case "usage_percent":
if want := "commandcode"; strings.Join(metric.Providers, ",") != want {
t.Fatalf("usage_percent providers = %v, want %s", metric.Providers, want)
}
}
}

if len(summary.Errors) != 0 {
t.Fatalf("errors = %+v, want none for a store that parses", summary.Errors)
}
AssertStoreUnchanged(t, req)
}
```
