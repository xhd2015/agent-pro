## Expected
- The command succeeds.
- stdout contains an opencode tool_use event for grep with mocked output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"tool_use"`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_grep_mock_ae"`)
	assertContains(t, resp.Stdout, `"tool":"grep"`)
	assertContains(t, resp.Stdout, `"status":"completed"`)
	assertContains(t, resp.Stdout, `mocked grep result`)
}
```
