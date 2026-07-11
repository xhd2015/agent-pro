## Expected

FetchStatus must auto-Skip the update menu and return usage:

1. `FetchOK` is true; `FetchError` empty.
2. `MonthlyUsage` is `58%` (fake 42% left).
3. `CreditsUsed` is `6519` (comma-stripped by ParseStatusSnapshot).
4. `CreditsTotal` is `11250`.
5. `NextReset` is `08:00 on 1 Aug`.
6. Marker `enter-update-now` is **absent**.

## Errors

- Timeout waiting for prompt/status (no auto-Skip / banner still blocks).
- Silent upgrade (`enter-update-now` marker).

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range resp.MarkerFiles {
		if name == "enter-update-now" {
			t.Fatalf("production Entered while Update now selected (marker %s)",
				filepath.Join(req.MarkerDir, name))
		}
	}
	if !resp.FetchOK {
		t.Fatalf("FetchOK=false error=%q, want ready usage after auto-Skip",
			resp.FetchError)
	}
	if strings.TrimSpace(resp.FetchError) != "" {
		t.Fatalf("FetchError=%q, want empty", resp.FetchError)
	}
	if resp.MonthlyUsage != "58%" {
		t.Fatalf("monthly_usage=%q, want 58%%", resp.MonthlyUsage)
	}
	if resp.CreditsUsed != "6519" {
		t.Fatalf("credits_used=%q, want 6519 (comma-stripped)", resp.CreditsUsed)
	}
	if resp.CreditsTotal != "11250" {
		t.Fatalf("credits_total=%q, want 11250", resp.CreditsTotal)
	}
	if resp.NextReset != "08:00 on 1 Aug" {
		t.Fatalf("next_reset=%q, want 08:00 on 1 Aug", resp.NextReset)
	}
}
```
