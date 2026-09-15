## Expected

- Exit 0: Command Code answered, so the run counts as a success even though two
  providers failed.
- stderr carries one `warning: <provider>: <error>` line per failed provider, so
  the failure is visible without being fatal.
- The human table names each provider in provider order, shows `(failed)` for grok
  and codex, and prints each provider's session total.
- All three snapshot files are written, including for the failed providers: the
  session block is real data even when the usage fetch is not.
- The failed records carry `ok:false` with the error and no usage values, while
  the Command Code record carries its values. The session totals stay as seeded.

## Side Effects

- Creates `$AGENT_PRO_HOME/usages/<provider>/<today>/`, three files.

## Errors

- Reported on stderr and inside the records; the exit status stays 0.

```go
import (
"regexp"
"strings"
"testing"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertExitCode(t, resp, 0)

AssertStderrContains(t, resp, "warning: grok: ")
AssertStderrContains(t, resp, "warning: codex: ")
if strings.Contains(resp.Stderr, "warning: commandcode: ") {
t.Fatalf("commandcode warned on a healthy fetch:\n%s", resp.Stderr)
}

if len(resp.Files) != 3 {
t.Fatalf("snapshot files = %v, want one per provider", resp.Files)
}
AssertStdoutContains(t, resp, "3 snapshots appended under ")
failedRow := regexp.MustCompile(`^(\w+)\s+\(failed\)\s+1\s+\S+$`)
for _, provider := range []usage.ProviderID{usage.Grok, usage.Codex} {
line := tableRow(t, resp, string(provider))
match := failedRow.FindStringSubmatch(line)
if match == nil {
t.Fatalf("%s row = %q, want \"<provider> (failed) 1 <took>\"", provider, line)
}
if got, want := match[1], string(provider); got != want {
t.Fatalf("row provider = %q, want %q", got, want)
}
}
if line := tableRow(t, resp, string(usage.CommandCode)); !strings.Contains(line, "25% used · Go") {
t.Fatalf("commandcode row = %q, want the projected headline", line)
}

for _, provider := range []usage.ProviderID{usage.Grok, usage.Codex} {
records := resp.RecordsFor(t, provider)
if len(records) != 1 {
t.Fatalf("%s wrote %d records, want 1", provider, len(records))
}
record := records[0]
if record.Usage.OK {
t.Fatalf("%s recorded a successful fetch", provider)
}
if !strings.Contains(record.Usage.Error, "fixture ") {
t.Fatalf("%s error = %q, want a missing fixture", provider, record.Usage.Error)
}
if len(record.Usage.Values) != 0 {
t.Fatalf("%s recorded values for a failed fetch: %v", provider, record.Usage.Values)
}
if got, want := record.Sessions.Total, 1; got != want {
t.Fatalf("%s sessions total = %d, want %d", provider, got, want)
}
}

records := resp.RecordsFor(t, usage.CommandCode)
if len(records) != 1 {
t.Fatalf("commandcode wrote %d records, want 1", len(records))
}
if !records[0].Usage.OK {
t.Fatalf("commandcode usage failed: %s", records[0].Usage.Error)
}
if got, want := records[0].Sessions.Total, 1; got != want {
t.Fatalf("commandcode sessions total = %d, want %d", got, want)
}
}

// tableRow returns the table line for one provider, without its detail line.
func tableRow(t *testing.T, resp *Response, provider string) string {
t.Helper()
for _, line := range resp.StdoutLines() {
if strings.HasPrefix(line, provider) {
return line
}
}
t.Fatalf("no table row for %s in:\n%s", provider, resp.Stdout)
return ""
}
```
