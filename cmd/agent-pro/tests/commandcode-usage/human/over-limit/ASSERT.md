## Expected

- ` 99% used` on a bar of 29 filled and 1 empty block: a sub-100 percentage
  never fills the whole bar even when rounding would.
- `Cycle: $0.06 left · 77 requests · 5 days to renewal`.
- Both window meters report ` 100%`; only the weekly one shows its reset,
  because the 5-hour window has no pending reset.
- Meter lines carry label, bar, percent, and reset only; `exceeded` is not
  annotated.

## Errors

- None.

```go
import (
"strings"
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
AssertSuccess(t, resp)

want := strings.Join([]string{
" USAGE  Go Plan · active",
"",
Bar(29, 30) + " 99% used",
"Cycle: $0.06 left · 77 requests · 5 days to renewal",
"",
"Usage limits",
"5-hour  " + Bar(30, 30) + " 100%",
"",
"Weekly  " + Bar(30, 30) + " 100% · resets in 2h 30m",
"",
"Full breakdown at commandcode.ai/alice/settings/usage",
}, "\n")
if got := strings.TrimRight(resp.Output, "\n"); got != want {
t.Fatalf("overlay mismatch\n got:\n%s\nwant:\n%s", got, want)
}

if got := resp.View.Credits.UsagePercent; got < 99.3 || got > 99.5 {
t.Fatalf("UsagePercent=%v, want 99.4", got)
}
if filled, empty := BarParts(FindLine(t, resp.Output, "% used")); filled != 29 || empty != 1 {
t.Fatalf("usage bar=%d filled / %d empty, want 29/1", filled, empty)
}
for _, label := range []string{"5-hour", "Weekly"} {
if filled, empty := BarParts(FindLine(t, resp.Output, label)); filled != 30 || empty != 0 {
t.Fatalf("%s bar=%d filled / %d empty, want 30/0", label, filled, empty)
}
}
AssertAbsent(t, resp.Output, "exceeded")
}
```
