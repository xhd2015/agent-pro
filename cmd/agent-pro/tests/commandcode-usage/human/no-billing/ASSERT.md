## Expected

- Header is the bare ` USAGE ` badge: no plan and no status are known.
- Body is `No billing data found.` followed by `Visit Studio for usage details.`
- No `Full breakdown` line: it belongs to the loaded body only.

## Errors

- One warning per failed endpoint, in fetch order: credits, subscriptions,
  then summary. None of them aborts the fetch.

```go
import (
"strings"
"testing"

"github.com/xhd2015/agent-pro/agent/commandcode"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
AssertSuccess(t, resp)
AssertAbsent(t, resp.Output, "\x1b")

want := strings.Join([]string{
" USAGE ",
"",
"No billing data found.",
"Visit Studio for usage details.",
}, "\n")
if got := strings.TrimRight(resp.Output, "\n"); got != want {
t.Fatalf("overlay mismatch\n got:\n%s\nwant:\n%s", got, want)
}

if resp.View.HasBillingData {
t.Fatalf("HasBillingData=true with no billing endpoints")
}
if resp.Data.Credits != nil || resp.Data.Subscription != nil || resp.Data.Summary != nil {
t.Fatalf("expected nil endpoint payloads, got %+v", resp.Data)
}
if len(resp.Data.Errors) != 3 {
t.Fatalf("Errors=%v, want one per failed endpoint", resp.Data.Errors)
}
joined := strings.Join(resp.Data.Errors, "\n")
for _, path := range []string{commandcode.PathCredits, commandcode.PathSubscriptions, commandcode.PathUsageSummary} {
if !strings.Contains(joined, path) {
t.Fatalf("no warning for %s:\n%s", path, joined)
}
}
AssertAbsent(t, resp.Output, "Full breakdown")
}
```
