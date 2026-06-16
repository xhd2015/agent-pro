## Expected
- Empty JSON array — ActionStepStart produces no grok events.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp.Output != `[]` && resp.Output != "null" && resp.Output != "" {
		t.Fatalf("expected empty result for ActionStepStart, got: %s", resp.Output)
	}
}
```
