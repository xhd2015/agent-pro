## Expected

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	if !strings.Contains(resp.Stdout, "terminal:  new iTerm2 window") {
		t.Fatalf("want new window:\n%s", resp.Stdout)
	}
	if resp.RunForegroundN != 0 || len(resp.Opened) != 0 {
		t.Fatalf("dry-run must not launch")
	}
}
```
