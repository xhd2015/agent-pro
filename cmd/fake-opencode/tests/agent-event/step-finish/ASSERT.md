## Expected
- The command succeeds.
- stdout contains an opencode `step_finish` event with session ID and a `part` of type `step-finish`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"step_finish"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_sf"`)
	assertContains(t, resp.Stdout, `"step-finish"`)
}
```
