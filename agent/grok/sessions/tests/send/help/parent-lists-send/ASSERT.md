## Expected

- Parent help line names send and session sources.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	out := resp.Stdout
	for _, want := range []string{"send", "--session-id", "--tab", "--open"} {
		if !strings.Contains(out, want) {
			t.Fatalf("parent help missing %q:\n%s", want, out)
		}
	}
	assertNoSend(t, resp)
}
```
