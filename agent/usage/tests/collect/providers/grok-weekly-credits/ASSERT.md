## Expected

- grok still reports usage: a weekly credit period rather than the monthly
  allowance.
- The percentage is the credit share (`2% used`, `98% remaining`) and the display
  period is `weekly`, so the snapshot says which kind of period the numbers
  belong to.
- No `monthly_limit`: only the monthly billing payload carries one.
- The record's session block is unaffected by which usage endpoint answered.

## Side Effects

- None.

## Errors

- None: the credits payload is a fallback, not a failure.

```go
import (
"testing"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

grok := resp.Record(t, usage.Grok)
AssertUsageOK(t, grok)

if got, want := Value(t, grok, "used_percent"), float64(2); got != want {
t.Fatalf("grok used_percent = %v, want %v", got, want)
}
if got, want := Value(t, grok, "remaining_percent"), float64(98); got != want {
t.Fatalf("grok remaining_percent = %v, want %v", got, want)
}
if got, want := Display(t, grok, "period"), "weekly"; got != want {
t.Fatalf("grok period = %q, want %q", got, want)
}
if got, want := Display(t, grok, "usage"), "2% used · weekly"; got != want {
t.Fatalf("grok usage headline = %q, want %q", got, want)
}
if _, ok := grok.Usage.Values["monthly_limit"]; ok {
t.Fatalf("weekly credits reported a monthly limit: %v", grok.Usage.Values)
}
AssertSessions(t, grok, 1, "2026-09-10T09:00:00Z", "2026-09-10T10:00:00Z")
}
```
