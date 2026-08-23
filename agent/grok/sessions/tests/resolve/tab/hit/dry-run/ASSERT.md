## Expected

- Dry-run plan includes mode tab, window, tab index, tty, and would resolve.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	out := resp.Stdout
	for _, need := range []string{
		"[dry-run] mode:          tab",
		"[dry-run] window:        100",
		"[dry-run] tab index:     2",
		"[dry-run] tty:           /dev/ttys102",
		"[dry-run] would resolve: " + fixtureTabSessionID,
		"[dry-run] confidence:    hard",
	} {
		if !strings.Contains(out, need) {
			t.Fatalf("dry-run stdout missing %q:\n%s", need, out)
		}
	}
	if strings.TrimSpace(out) == fixtureTabSessionID {
		t.Fatal("dry-run must not emit bare-id-only success shape")
	}
}
```
