## Expected

- `endpoint` is `usage.EndpointCommandCodeUsage`: three endpoints feed one
  snapshot, so the label names the composite fetch rather than one path.
- Credits: `usage_percent` 25 (7.5 of the 10 credit allowance left), 7.5 credits
  remaining, 2.5 credits spent.
- Window limits become percentages: `five_hour_used_percent` 50 (1.5 of 3) and
  `weekly_used_percent` 25 (1.5 of 6), with the weekly reset timestamp in
  milliseconds as the API reports it.
- Summary counters: 120 requests, 5 failed, 600 tokens in and 400 out.
- `days_to_renewal` 3, measured from the injected clock to the period end.
- Display keeps the provider's own strings: plan `Go`, headline `25% used · Go`,
  account email, and the account's usage URL.

## Side Effects

- None.

## Errors

- None.

```go
import (
"testing"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

cc := resp.Record(t, usage.CommandCode)
AssertUsageOK(t, cc)

if got, want := cc.Usage.Endpoint, usage.EndpointCommandCodeUsage; got != want {
t.Fatalf("endpoint = %q, want %q", got, want)
}
AssertValues(t, cc, map[string]float64{
"usage_percent":          25,
"credits_remaining":      7.5,
"cost_usd":               2.5,
"five_hour_used_percent": 50,
"weekly_used_percent":    25,
"weekly_reset_at":        1789600000000,
"requests_total":         120,
"requests_failed_total":  5,
"tokens_in_total":        600,
"tokens_out_total":       400,
"days_to_renewal":        3,
})
if got, want := Display(t, cc, "plan"), "Go"; got != want {
t.Fatalf("plan = %q, want %q", got, want)
}
if got, want := Display(t, cc, "usage"), "25% used · Go"; got != want {
t.Fatalf("usage headline = %q, want %q", got, want)
}
if got, want := Display(t, cc, "email"), "alice@example.test"; got != want {
t.Fatalf("email = %q, want %q", got, want)
}
if got := Display(t, cc, "usage_url"); got == "" || got[len(got)-len("/alice/settings/usage"):] != "/alice/settings/usage" {
t.Fatalf("usage_url = %q, want it to end with /alice/settings/usage", got)
}
}
```
