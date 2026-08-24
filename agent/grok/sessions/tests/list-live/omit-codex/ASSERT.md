## Expected

- Codex runner does not produce a list-live row.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if strings.Contains(resp.Stdout, fixtureListLiveSID) {
		t.Fatalf("codex must not appear in grok list-live:\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stdout, "0 sessions", "footer")
}
```
