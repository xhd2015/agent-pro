## Expected

- The overrides reach every scoped endpoint: `credits`, `subscriptions`, and
  `usage/summary` all carry `orgId=org_pinned`, not the `org_team` whoami
  reported.
- `usage/summary` carries `since=2026-01-01T00:00:00Z`, not the subscription's
  own period start.
- Pinning the period start does not move the renewal date: the cycle line
  still derives from `currentPeriodEnd` in the response.
- whoami is unaffected by `--org`: it stays personal (`limits=1`).

## Errors

- None.

```go
import (
"strings"
"testing"

"github.com/xhd2015/agent-pro/agent/commandcode"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
AssertSuccess(t, resp)

for _, path := range []string{commandcode.PathCredits, commandcode.PathSubscriptions, commandcode.PathUsageSummary} {
if got := ParseQuery(t, req.Queries[path]).Get("orgId"); got != "org_pinned" {
t.Fatalf("%s orgId=%q, want the pinned org_pinned", path, got)
}
}
if got := ParseQuery(t, req.Queries[commandcode.PathWhoami]).Get("orgId"); got != "" {
t.Fatalf("whoami orgId=%q, want it omitted", got)
}

summaryQuery := ParseQuery(t, req.Queries[commandcode.PathUsageSummary])
if got := summaryQuery.Get("since"); got != "2026-01-01T00:00:00Z" {
t.Fatalf("summary since=%q, want the pinned 2026-01-01T00:00:00Z", got)
}
if resp.Data.Subscription == nil {
t.Fatalf("subscription did not load")
}
if derived := resp.Data.Subscription.CurrentPeriodStart; derived == summaryQuery.Get("since") {
t.Fatalf("subscription period start equals the pinned since (%q): the fixture cannot prove the override", derived)
}
if got := resp.Data.Summary; got == nil || got.PeriodBasis != "custom" {
t.Fatalf("summary=%+v, want the custom-period payload", got)
}

AssertContains(t, resp.Output, "Cycle: $3.25 left · 7 requests · 20 days to renewal")
if !strings.Contains(resp.Output, "92% used") {
t.Fatalf("missing the 92%% usage marker in:\n%s", resp.Output)
}
}
```
