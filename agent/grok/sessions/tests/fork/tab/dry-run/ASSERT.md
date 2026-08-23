## Expected

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	for _, need := range []string{
		"Would fork grok session",
		"window:     100",
		"tab index:  2",
		"tty:        /dev/ttys102",
		"grok id:   " + fixtureTabSessionID,
		"terminal:  current",
	} {
		if !strings.Contains(resp.Stdout, need) {
			t.Fatalf("dry-run missing %q:\n%s", need, resp.Stdout)
		}
	}
	if resp.RunForegroundN != 0 || len(resp.Opened) != 0 {
		t.Fatal("dry-run must not launch")
	}
}
```
