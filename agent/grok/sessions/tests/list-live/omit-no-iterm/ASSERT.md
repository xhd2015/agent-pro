## Expected

- Session id omitted; footer `0 sessions`.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if strings.Contains(resp.Stdout, fixtureListLiveSID) {
		t.Fatalf("sid without iTerm must be omitted:\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stdout, "0 sessions", "footer")
}
```
