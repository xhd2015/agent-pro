## Expected

- Empty cwd without `--dir` is fatal; no open.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "empty cwd") {
		t.Fatalf("error = %v, want empty cwd", resp.Err)
	}
	assertNoSideEffects(t, resp)
}
```
