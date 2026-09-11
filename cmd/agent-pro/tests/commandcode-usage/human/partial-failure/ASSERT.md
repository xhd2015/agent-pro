## Expected

- The plan still comes from the subscription: ` USAGE  Go Plan · active`.
- With no credit balance, the bar is replaced by `Plan details unavailable`
  and the `Cycle:` line is omitted.
- No `Usage limits` section: window limits ride on the credits payload.
- `Full breakdown at commandcode.ai/alice/settings/usage` still renders.

## Errors

- Two warnings: an unauthorized credits response maps to
  `Session expired`, and the summary failure keeps its endpoint and status.

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
"Plan details unavailable",
"",
"Full breakdown at commandcode.ai/alice/settings/usage",
}, "\n")
if got := strings.TrimRight(resp.Output, "\n"); got != want {
t.Fatalf("overlay mismatch\n got:\n%s\nwant:\n%s", got, want)
}
AssertAbsent(t, resp.Output, "Cycle:")
AssertAbsent(t, resp.Output, "Usage limits")

if !resp.View.HasBillingData {
t.Fatalf("HasBillingData=false with a loaded subscription")
}
if resp.View.Plan == nil || resp.View.Plan.Name != "Go" {
t.Fatalf("plan=%+v", resp.View.Plan)
}
if resp.Data.Credits != nil {
t.Fatalf("Credits=%+v, want nil", resp.Data.Credits)
}
if got := resp.View.Credits.TotalRemaining; got != 0 {
t.Fatalf("TotalRemaining=%v, want 0", got)
}
if resp.View.WindowLimits != nil {
t.Fatalf("WindowLimits=%+v, want nil", resp.View.WindowLimits)
}

if len(resp.Data.Errors) != 2 {
t.Fatalf("Errors=%v, want 2", resp.Data.Errors)
}
if resp.Data.Errors[0] != commandcode.ErrSessionExpired {
t.Fatalf("Errors[0]=%q, want %q", resp.Data.Errors[0], commandcode.ErrSessionExpired)
}
summary := resp.Data.Errors[1]
if !strings.Contains(summary, commandcode.PathUsageSummary) || !strings.Contains(summary, "502") {
t.Fatalf("Errors[1]=%q, want the summary endpoint and status", summary)
}
}
```
