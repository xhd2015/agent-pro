## Expected

- SoftExit called once.
- OpenFn used (production open launch).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "OpenCalls", resp.OpenCalls, 1)
	assertEqual(t, "SoftExitCalls", resp.SoftExitCalls, 1)
}
```
