## Expected

- Parent help line names `list-live`.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if !strings.Contains(resp.Stdout, "list-live") {
		t.Fatalf("parent help missing list-live:\n%s", resp.Stdout)
	}
}
```
