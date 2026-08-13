## Expected

- API error mentions prompt.
- Launch not called.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertAPIError(t, resp)
	assertContains(t, resp.ErrString, "prompt")
	assertEqual(t, "LaunchCalls", resp.LaunchCalls, 0)
}
```
