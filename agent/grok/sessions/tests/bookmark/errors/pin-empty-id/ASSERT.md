## Expected

- Non-nil error (empty id rejected or not found).
- Store file still absent.

## Errors

- any error; store not created

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	assertPathMissing(t, storePath(req.AgentProHome))
}
```
