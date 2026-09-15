## Expected

- Exit 1, with `agent-pro: no usage collected: every provider failed` on stderr, so
  a scheduler alerts instead of silently recording nothing.
- One `warning:` line per provider, before the final error: the person reading the
  log sees which endpoint failed, not just that everything did.
- stdout still carries the three JSON records, each with `ok:false` and its error:
  `--json` is the run's record stream, and the exit status is what marks the run as
  failed, so a pipeline that captures the records also captures why they are empty.
- The three snapshot files are still written, each with `ok:false` and its error,
  and the grok record still carries the session count from its home. History keeps
  a trace of the outage, and the session trend does not break.

## Side Effects

- Creates `$AGENT_PRO_HOME/usages/<provider>/<today>/`, three files.

## Errors

- The CLI reports the failure on stderr and through the exit status; per-provider
  errors are in the records.

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
AssertExitCode(t, resp, 1)

AssertStderrContains(t, resp, "agent-pro: no usage collected: every provider failed")
for _, provider := range []string{"grok", "codex", "commandcode"} {
AssertStderrContains(t, resp, "warning: "+provider+": ")
}

if strings.TrimSpace(resp.Stdout) == "" {
t.Fatal("stdout is empty, want one record per provider even when every fetch failed")
}
printed := DecodeJSONLines(t, resp.Stdout)
if len(printed) != 3 {
t.Fatalf("stdout has %d records, want 3:\n%s", len(printed), resp.Stdout)
}
for _, record := range printed {
if record.Usage.OK {
t.Fatalf("%s printed a successful record", record.Provider)
}
}

if len(resp.Files) != 3 {
t.Fatalf("snapshot files = %v, want one per provider even on failure", resp.Files)
}
for _, provider := range []usage.ProviderID{usage.Grok, usage.Codex, usage.CommandCode} {
records := resp.RecordsFor(t, provider)
if len(records) != 1 {
t.Fatalf("%s wrote %d records, want 1", provider, len(records))
}
if records[0].Usage.OK {
t.Fatalf("%s recorded a successful fetch", provider)
}
if records[0].Usage.Error == "" {
t.Fatalf("%s recorded no error", provider)
}
}

grok := resp.RecordsFor(t, usage.Grok)[0]
if got, want := grok.Sessions.Total, 1; got != want {
t.Fatalf("grok sessions total = %d, want %d from its home", got, want)
}
cc := resp.RecordsFor(t, usage.CommandCode)[0]
if !strings.Contains(cc.Usage.Error, "not authenticated") {
t.Fatalf("commandcode error = %q, want the missing-credential message", cc.Usage.Error)
}
}
```
