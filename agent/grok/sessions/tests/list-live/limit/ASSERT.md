## Expected

- Exactly one session row in the footer after `--limit 1`.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertContains(t, resp.Stdout, "1 sessions", "footer")
	// Sorted: aaaa… comes before bbbb…
	if !strings.Contains(resp.Stdout, fixtureListLiveSID) {
		t.Fatalf("want first sorted sid:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, fixtureListLiveSID2) {
		t.Fatalf("second sid must be limited out:\n%s", resp.Stdout)
	}
}
```
