## Expected

- `--index` cannot combine with `--tab`.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "--index") {
		t.Fatalf("error = %v, want --index combine error", resp.Err)
	}
	assertNoContents(t, resp)
}
```
