## Expected

- Positional id + `--tab` is fatal.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "cannot be combined") {
		t.Fatalf("error = %v", resp.Err)
	}
	assertNoSideEffects(t, resp)
}
```
