## Expected

- `--help` documents session id, `--tab`, `--tab-index`, `--index`, and resume-when-missing.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	out := resp.Stdout
	for _, want := range []string{"<session-id>", "--tab", "--tab-index", "--index", "grok --resume", "--dry-run"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
	assertNoSideEffects(t, resp)
}
```
