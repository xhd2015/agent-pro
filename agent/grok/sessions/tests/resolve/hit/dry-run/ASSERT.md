## Expected

- Plan lines with `[dry-run]` prefix; no bare-id-only stdout; stderr empty.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	assertStdoutExact(t, resp.Stdout,
		"[dry-run] start pid:     6000",
		"[dry-run] ancestor pid:  4242",
		"[dry-run] runner pid:    4242",
		"[dry-run] would resolve: "+fixtureSessionID,
		"[dry-run] source:        open-files",
		"[dry-run] confidence:    hard",
	)
	if resp.Stderr != "" {
		t.Fatalf("stderr want empty, got %q", resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) == fixtureSessionID {
		t.Fatal("dry-run must not use bare-id success shape")
	}
}
```
