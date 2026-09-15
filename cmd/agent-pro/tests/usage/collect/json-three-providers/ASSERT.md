## Expected

- Exit 0: at least one provider answered.
- stdout is exactly three JSON lines in provider order `grok`, `codex`,
  `commandcode`, so a pipeline can consume the records without parsing a table.
- There is no warning on stderr: nothing failed, so nothing competes with the
  records on the machine-readable stream.
- Every record carries schema `agent-pro/usage-snapshot/v1` and the run's
  timestamp, close to now because `collect` stamps the real clock.
- grok reports the billing endpoint with 73% used, codex 62% used, and Command
  Code the composite endpoint; each record also reports the sessions found in its
  home.
- Three files exist under `<home>/usages`, one per provider directory, in today's
  local day directory, named `<HH-MM-SS>-snapshot.jsonl`, each holding exactly the
  record that was printed.

## Side Effects

- Creates `$AGENT_PRO_HOME/usages/<provider>/<today>/`.

## Errors

- None.

```go
import (
"path/filepath"
"reflect"
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
AssertExitCode(t, resp, 0)

printed := DecodeJSONLines(t, resp.Stdout)
if got, want := providersOf(printed), []string{"grok", "codex", "commandcode"}; !reflect.DeepEqual(got, want) {
t.Fatalf("printed providers = %v, want %v", got, want)
}
for _, record := range printed {
if record.Schema != usage.SchemaID {
t.Fatalf("%s schema = %q, want %q", record.Provider, record.Schema, usage.SchemaID)
}
if delta := time.Since(record.TS).Abs(); delta > 2*time.Minute {
t.Fatalf("%s ts = %s, want within 2m of now", record.Provider, record.TS)
}
if !record.Usage.OK {
t.Fatalf("%s usage not ok: %s", record.Provider, record.Usage.Error)
}
}

grok := findPrinted(t, printed, usage.Grok)
if got, want := grok.Usage.Endpoint, "billing"; got != want {
t.Fatalf("grok endpoint = %q, want %q", got, want)
}
if got, want := grok.Usage.Values["used_percent"], float64(73); got != want {
t.Fatalf("grok used_percent = %v, want %v", got, want)
}
if got, want := grok.Sessions.Total, 1; got != want {
t.Fatalf("grok sessions total = %d, want %d", got, want)
}

codex := findPrinted(t, printed, usage.Codex)
if got, want := codex.Usage.Values["used_percent"], float64(62); got != want {
t.Fatalf("codex used_percent = %v, want %v", got, want)
}
if got, want := codex.Sessions.Total, 1; got != want {
t.Fatalf("codex sessions total = %d, want %d", got, want)
}

cc := findPrinted(t, printed, usage.CommandCode)
if got, want := cc.Usage.Endpoint, usage.EndpointCommandCodeUsage; got != want {
t.Fatalf("commandcode endpoint = %q, want %q", got, want)
}
if got, want := cc.Sessions.Total, 1; got != want {
t.Fatalf("commandcode sessions total = %d, want %d", got, want)
}

if resp.Stderr != "" {
t.Fatalf("stderr = %q, want empty on a clean run", resp.Stderr)
}

if len(resp.Files) != 3 {
t.Fatalf("snapshot files = %v, want one per provider", resp.Files)
}
name := regexp.MustCompile(`^\d{2}-\d{2}-\d{2}-snapshot\.jsonl$`)
for _, path := range resp.Files {
if !name.MatchString(filepath.Base(path)) {
t.Fatalf("snapshot file %s is not named HH-MM-SS-snapshot.jsonl", path)
}
provider := usage.ProviderID(filepath.Base(filepath.Dir(filepath.Dir(path))))
AssertStoreDay(t, req.Home, provider, path)
lines := resp.Lines[path]
if len(lines) != 1 {
t.Fatalf("%s holds %d lines, want 1", path, len(lines))
}
stored := DecodeRecord(t, lines[0])
want := findPrinted(t, printed, provider)
if !stored.TS.Equal(want.TS) || stored.Provider != want.Provider {
t.Fatalf("stored %s record = %s/%s, want the printed %s/%s",
provider, stored.TS, stored.Provider, want.TS, want.Provider)
}
if !strings.Contains(string(resp.Stdout), lines[0]) {
t.Fatalf("stdout does not carry the stored record:\n%s", resp.Stdout)
}
}
}

func providersOf(records []usage.Record) []string {
out := make([]string, 0, len(records))
for _, record := range records {
out = append(out, string(record.Provider))
}
return out
}

func findPrinted(t *testing.T, records []usage.Record, provider usage.ProviderID) usage.Record {
t.Helper()
for _, record := range records {
if record.Provider == provider {
return record
}
}
t.Fatalf("no printed record for %s", provider)
return usage.Record{}
}
```
