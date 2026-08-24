## Expected

- ITERM shows `(+1)` for the second hosting tab.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if !strings.Contains(resp.Stdout, "(+1)") {
		t.Fatalf("want multi-tab (+1); got:\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stdout, fixtureListLiveSID, "session id")
}
```
