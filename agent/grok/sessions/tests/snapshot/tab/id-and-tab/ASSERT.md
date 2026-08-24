## Expected

- Positional id + `--tab` is a usage error.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "--tab") {
		t.Fatalf("error = %v, want --tab combine error", resp.Err)
	}
	assertNoContents(t, resp)
}
```
