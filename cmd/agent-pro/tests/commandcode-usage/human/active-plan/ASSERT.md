## Expected

- Header ` USAGE  Go Plan · active`.
- Bar of 23 filled and 7 empty blocks with ` 76% used` (75.914% rounded).
- `Cycle: $2.41 left · 1,040 requests · 5 days to renewal`.
- `Usage limits` with `5-hour` at 0% and `Weekly` at 8%, resetting in 7h 4m.
- `Full breakdown at commandcode.ai/alice/settings/usage`.
- No ANSI escapes: color defaults off.

## Errors

- None.

```go
import (
"strings"
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
AssertSuccess(t, resp)
AssertAbsent(t, resp.Output, "\x1b")

want := strings.Join([]string{
" USAGE  Go Plan · active",
"",
Bar(23, 30) + " 76% used",
"Cycle: $2.41 left · 1,040 requests · 5 days to renewal",
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

if filled, empty := BarParts(FindLine(t, resp.Output, "% used")); filled != 23 || empty != 7 {
t.Fatalf("bar=%d filled / %d empty, want 23/7", filled, empty)
}
if resp.View.Plan == nil || resp.View.Plan.Name != "Go" {
t.Fatalf("plan=%+v", resp.View.Plan)
}
if got := resp.View.Credits.UsagePercent; got < 75.9 || got > 75.95 {
t.Fatalf("UsagePercent=%v", got)
}
if got := resp.View.Credits.TotalPool; got != 10 {
t.Fatalf("TotalPool=%v", got)
}
if resp.View.DaysLeft == nil || *resp.View.DaysLeft != 5 {
t.Fatalf("DaysLeft=%v", resp.View.DaysLeft)
}
}
```
