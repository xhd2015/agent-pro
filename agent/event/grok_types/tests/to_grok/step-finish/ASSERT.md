## Expected
- Empty JSON array — ActionStepFinish produces no grok events.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp.Output != `[]` && resp.Output != "null" && resp.Output != "" {
		t.Fatalf("expected empty result for ActionStepFinish, got: %s", resp.Output)
	}
}
```
