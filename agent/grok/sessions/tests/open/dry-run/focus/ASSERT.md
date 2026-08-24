## Expected

- Dry-run focus prints plan; never focuses or opens.

## Expected Output

```text
Would focus: window 3, tab 1
```

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertNoSideEffects(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
---
Would focus: window 3, tab 1
`)
}
```
