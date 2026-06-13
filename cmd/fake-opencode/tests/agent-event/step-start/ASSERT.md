## Expected
- The command succeeds.
- stdout contains an opencode `step_start` event with session ID and a `part` of type `step-start`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"step_start"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_ss"`)
	assertContains(t, resp.Stdout, `"step-start"`)
}
```
