## Expected

- codex reports usage with `wham/usage` as the endpoint that answered, so a
  stored snapshot shows which API produced the numbers.
- The values and display strings are the same as the primary endpoint's: 62% used,
  38% remaining, plan `business`, resets 2026-09-01T00:00:00Z.
- The record's session block is unaffected.

## Side Effects

- None.

## Errors

- None: the wham endpoint is a fallback, not a failure.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

codex := resp.Record(t, usage.Codex)
AssertUsageOK(t, codex)

if got, want := codex.Usage.Endpoint, "wham/usage"; got != want {
t.Fatalf("codex endpoint = %q, want %q", got, want)
}
AssertValues(t, codex, map[string]float64{
"used_percent":      62,
"remaining_percent": 38,
"reset_at":          float64(1788220800),
})
if got, want := Display(t, codex, "usage"), "62% used"; got != want {
t.Fatalf("codex usage headline = %q, want %q", got, want)
}
if got, want := Display(t, codex, "plan_type"), "business"; got != want {
t.Fatalf("codex plan type = %q, want %q", got, want)
}
AssertSessions(t, codex, 1,
time.Date(2026, 9, 10, 9, 0, 0, 0, time.Local).UTC().Format(time.RFC3339),
req.Now.Add(-24*time.Hour).UTC().Format(time.RFC3339))
}
```
