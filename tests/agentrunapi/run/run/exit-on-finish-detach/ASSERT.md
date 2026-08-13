## Expected

- SoftExit called once (detach default is exit).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "SoftExitCalls", resp.SoftExitCalls, 1)
}
```
