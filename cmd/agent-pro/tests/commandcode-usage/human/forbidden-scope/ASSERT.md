## Expected

- The overlay still renders from the billing data that did load, minus the
  summary: the `Cycle:` line keeps the balance and drops the request count.
- The `Usage limits` section and the Studio URL are unaffected.

## Errors

- One warning carrying the API's permission message and its status. Reporting
  this as `Session expired` would send the reader to re-authenticate a
  credential that is working.

```go
import (
"strings"
"testing"

"github.com/xhd2015/agent-pro/agent/commandcode"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
AssertSuccess(t, resp)

want := strings.Join([]string{
" USAGE  Go Plan · active",
"",
Bar(23, 30) + " 76% used",
"Cycle: $2.41 left · 5 days to renewal",
"",
"Usage limits",
"5-hour  " + Bar(0, 30) + " 0%",
"",
"Weekly  " + Bar(2, 30) + " 8% · resets in 7h 4m",
"",
"Full breakdown at commandcode.ai/alice/settings/usage",
}, "\n")
if got := strings.TrimRight(resp.Output, "\n"); got != want {
t.Fatalf("overlay mismatch\n got:\n%s\nwant:\n%s", got, want)
}
AssertAbsent(t, resp.Output, "requests")

if resp.Data.Summary != nil {
t.Fatalf("Summary=%+v, want nil", resp.Data.Summary)
}
if len(resp.Data.Errors) != 1 {
t.Fatalf("Errors=%v, want one summary failure", resp.Data.Errors)
}
msg := resp.Data.Errors[0]
if !strings.Contains(msg, "do not have permission") || !strings.Contains(msg, "403") {
t.Fatalf("Errors[0]=%q, want the API permission message and status", msg)
}
if strings.Contains(msg, commandcode.ErrSessionExpired) {
t.Fatalf("Errors[0]=%q, want a permission error, not an expired session", msg)
}
if got := strings.Count(msg, commandcode.PathUsageSummary); got != 1 {
t.Fatalf("Errors[0]=%q names the endpoint %d times, want 1", msg, got)
}
}
```
