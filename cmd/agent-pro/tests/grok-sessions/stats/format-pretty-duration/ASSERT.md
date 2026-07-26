## Expected

- Raw `SessionStats` latency fields stay numeric: 120 s, 1500 ms, 400 ms.
- Human output Latency lines use pretty forms:
  - Duration contains `2m` (not bare `120s` as the only form)
  - Avg response contains `1.5s`
  - Avg time-to-first contains `400ms`

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	st := resp.Stats
	if st.SessionDurationSec != 120 || st.AvgResponseMs != 1500 || st.AvgTTFTMs != 400 {
		t.Fatalf("raw latency = %ds / %dms / %dms, want 120/1500/400",
			st.SessionDurationSec, st.AvgResponseMs, st.AvgTTFTMs)
	}

	out := resp.Output
	assertContains(t, out, "Latency")
	assertContains(t, out, "Duration:")
	assertContains(t, out, "Avg response:")
	assertContains(t, out, "Avg time-to-first:")

	// Pretty forms (RED until implementer formats durations).
	if !strings.Contains(out, "2m") {
		t.Fatalf("Duration should pretty-print 120s as 2m, got:\n%s", out)
	}
	if !strings.Contains(out, "1.5s") {
		t.Fatalf("Avg response should pretty-print 1500ms as 1.5s, got:\n%s", out)
	}
	if !strings.Contains(out, "400ms") {
		t.Fatalf("Avg time-to-first should show 400ms, got:\n%s", out)
	}
	// Reject only-raw second form for session duration when pretty is required.
	// Allow "2m" without requiring absence of every "s" token.
	if strings.Contains(out, "Duration: 120s") && !strings.Contains(out, "Duration: 2m") {
		t.Fatalf("Duration still raw 120s without pretty 2m:\n%s", out)
	}
}
```
