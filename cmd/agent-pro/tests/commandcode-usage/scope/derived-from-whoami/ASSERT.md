## Expected

- `whoami` is always personal: `limits=1` and no `orgId` parameter. Passing an
  org there is what the API rejects, and the identity lookup is what supplies
  the org in the first place.
- `credits`, `subscriptions`, and `usage/summary` all carry `orgId=org_team`
  from `whoami.org.id`.
- `usage/summary` carries `since` equal to the subscription's
  `currentPeriodStart`, so the totals cover the billing period.
- The org login names the Studio URL: `commandcode.ai/acme/settings/usage`.
- Every request carries the credentials as a bearer token, so the scoped calls
  are authenticated the same way the identity call is.

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

for _, path := range []string{commandcode.PathWhoami, commandcode.PathCredits, commandcode.PathSubscriptions, commandcode.PathUsageSummary} {
authz := req.AuthHeaders[path]
if !strings.HasPrefix(authz, "Bearer ") || strings.TrimPrefix(authz, "Bearer ") == "" {
t.Fatalf("%s Authorization=%q, want a bearer token", path, authz)
}
}

whoami := ParseQuery(t, req.Queries[commandcode.PathWhoami])
if got := whoami.Get("limits"); got != "1" {
t.Fatalf("whoami limits=%q, want 1", got)
}
if got := whoami.Get("orgId"); got != "" {
t.Fatalf("whoami orgId=%q, want it omitted", got)
}

for _, path := range []string{commandcode.PathCredits, commandcode.PathSubscriptions, commandcode.PathUsageSummary} {
if got := ParseQuery(t, req.Queries[path]).Get("orgId"); got != "org_team" {
t.Fatalf("%s orgId=%q, want org_team from whoami", path, got)
}
}

if resp.Data.Subscription == nil {
t.Fatalf("subscription did not load")
}
since := ParseQuery(t, req.Queries[commandcode.PathUsageSummary]).Get("since")
if want := resp.Data.Subscription.CurrentPeriodStart; since != want {
t.Fatalf("summary since=%q, want the subscription period start %q", since, want)
}

if got := resp.View.Plan; got == nil || got.ID != "teams-pro" || got.MonthlyCredits != 40 {
t.Fatalf("plan=%+v, want teams-pro with a 40-credit allowance", got)
}
if got := resp.View.UsageURLDisplay; got != "commandcode.ai/acme/settings/usage" {
t.Fatalf("UsageURLDisplay=%q, want the org login in the path", got)
}
AssertContains(t, resp.Output, "Teams Pro Plan")
}
```
