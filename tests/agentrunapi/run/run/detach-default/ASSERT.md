## Expected

- No API error.
- Launch called once with `OpenTerminal=false`.
- OpenFn not used (detach does not open iTerm).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "LaunchCalls", resp.LaunchCalls, 1)
	assertEqual(t, "LaunchOpenTerminal", resp.LaunchOpenTerminal, false)
	assertEqual(t, "OpenCalls", resp.OpenCalls, 0)
}
```
