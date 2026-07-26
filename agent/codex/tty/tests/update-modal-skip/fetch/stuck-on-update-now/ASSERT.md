---
label: slow && negative
explanation: stuck update menu; waits for error/timeout until early Skip failure
---

## Expected

1. `FetchOK` is false.
2. `FetchError` is non-empty (timeout or could not select Skip).
3. Markers `enter-while-update-now` / `enter-update-now` are **absent**.
4. `MonthlyUsage` is empty.

## Errors

- Success without Skip.
- Enter while Update now selected.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range resp.MarkerFiles {
		if name == "enter-while-update-now" || name == "enter-update-now" {
			t.Fatalf("production confirmed Update now (marker %q) — must verify Skip before Enter", name)
		}
	}
	if resp.FetchOK {
		t.Fatalf("FetchOK=true with stuck Update now fake — must not succeed without Skip")
	}
	if strings.TrimSpace(resp.FetchError) == "" {
		t.Fatal("FetchError empty, want timeout or could not select Skip")
	}
	if strings.TrimSpace(resp.MonthlyUsage) != "" {
		t.Fatalf("monthly_usage=%q, want empty on error", resp.MonthlyUsage)
	}
}
```
